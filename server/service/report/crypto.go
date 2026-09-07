package report

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// 数据源密码 AES-256-GCM 加解密。
// 密钥来自 config.yaml report.aes-key（32 字节随机数的 base64）；
// 密文自识别格式：enc:<base64(nonce(12B)+ciphertext+tag)>，支持未来换钥迁移。
const (
	passwordCipherPrefix = "enc:"
	passwordMaskedText   = "******"
)

// sourceConfigJSON 连接配置 JSON 结构 {dsn, username, password}（结构体编解码保持字段序）
type sourceConfigJSON struct {
	Dsn      string `json:"dsn"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ValidateAesKey 启动校验：密钥已配置但非法时报错提示（未配置仅告警，
// 数据源保存路径会以明确错误拒绝，保证明文密码不落库）
func ValidateAesKey() error {
	key := global.GVA_CONFIG.Report.AesKey
	if key == "" {
		return nil
	}
	_, err := aesKeyBytes()
	return err
}

// aesKeyBytes 解析配置密钥（base64 解码后须 32 字节）
func aesKeyBytes() ([]byte, error) {
	raw := global.GVA_CONFIG.Report.AesKey
	if raw == "" {
		return nil, errors.New("report.aes-key 未配置，数据源密码加密不可用（32字节随机数base64，生成：openssl rand -base64 32）")
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("report.aes-key 不是合法的 base64: %v", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("report.aes-key 解码后须为32字节，当前 %d 字节", len(key))
	}
	return key, nil
}

func aesGcmEncrypt(plain string) (string, error) {
	key, err := aesKeyBytes()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return passwordCipherPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

func aesGcmDecrypt(cipherText string) (string, error) {
	key, err := aesKeyBytes()
	if err != nil {
		return "", err
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(cipherText, passwordCipherPrefix))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("密文长度不合法")
	}
	plain, err := gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// parseSourceConfig 解析连接配置 JSON（仅校验格式，不校验 dsn 非空——http 类型无连接配置）
func parseSourceConfig(sourceConfig string) (*sourceConfigJSON, error) {
	if strings.TrimSpace(sourceConfig) == "" {
		return &sourceConfigJSON{}, nil
	}
	var cfg sourceConfigJSON
	if err := json.Unmarshal([]byte(sourceConfig), &cfg); err != nil {
		return nil, ErrSourceConfigInvalid
	}
	return &cfg, nil
}

func marshalSourceConfig(cfg *sourceConfigJSON) string {
	b, _ := json.Marshal(cfg)
	return string(b)
}

// EncryptSourceConfig 加密 sourceConfig 中的 password 字段（创建数据源用；其余字段原样保留）
func EncryptSourceConfig(sourceConfig string) (string, error) {
	cfg, err := parseSourceConfig(sourceConfig)
	if err != nil {
		return "", err
	}
	if cfg.Password == "" {
		return marshalSourceConfig(cfg), nil
	}
	sealed, err := aesGcmEncrypt(cfg.Password)
	if err != nil {
		return "", err
	}
	cfg.Password = sealed
	return marshalSourceConfig(cfg), nil
}

// DecryptSourceConfig 解密 sourceConfig 中的 password 字段（内部取数用，不出 API）
func DecryptSourceConfig(sourceConfig string) (string, error) {
	cfg, err := parseSourceConfig(sourceConfig)
	if err != nil {
		return "", err
	}
	if cfg.Password == "" {
		return marshalSourceConfig(cfg), nil
	}
	if !strings.HasPrefix(cfg.Password, passwordCipherPrefix) {
		// 兼容历史明文（正常路径不会出现），原样返回
		return marshalSourceConfig(cfg), nil
	}
	plain, err := aesGcmDecrypt(cfg.Password)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrSourceConfigInvalid, err)
	}
	cfg.Password = plain
	return marshalSourceConfig(cfg), nil
}

// MaskSourceConfig password 脱敏为 ******（供 RespVO 回显）
func MaskSourceConfig(sourceConfig string) string {
	cfg, err := parseSourceConfig(sourceConfig)
	if err != nil {
		return "{}"
	}
	if cfg.Password != "" {
		cfg.Password = passwordMaskedText
	}
	return marshalSourceConfig(cfg)
}

// HandlePasswordOnUpdate 编辑态密码三态处理：
// 提交 password 为空或 ****** → 保留库中原密文；提交新值 → 加密替换（dsn/username 以新提交为准）
func HandlePasswordOnUpdate(newSourceConfig, oldSourceConfig string) (string, error) {
	newCfg, err := parseSourceConfig(newSourceConfig)
	if err != nil {
		return "", err
	}
	if newCfg.Password != "" && newCfg.Password != passwordMaskedText {
		return EncryptSourceConfig(newSourceConfig)
	}
	oldCfg, err := parseSourceConfig(oldSourceConfig)
	if err != nil {
		return "", err
	}
	newCfg.Password = oldCfg.Password
	return marshalSourceConfig(newCfg), nil
}
