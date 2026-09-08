<template>
  <!-- select 形态：全量缓存 + 前端过滤 -->
  <el-select
    v-if="mode === 'select'"
    :model-value="modelValue"
    :multiple="multiple"
    filterable
    :disabled="disabled"
    :clearable="clearable"
    :placeholder="placeholder"
    :size="size"
    style="width: 100%"
    @update:model-value="(v) => emit('update:modelValue', v)"
    @change="onChange"
  >
    <el-option v-for="u in options" :key="u.ID" :label="formatUser(u)" :value="u.ID" />
  </el-select>

  <!-- dialog 形态：只读触发器 + 弹窗内芯（append-to-body 规避嵌套表单容器） -->
  <div
    v-else
    class="gva-choose-user-trigger"
    :class="[size && `is-${size}`, { 'is-disabled': disabled }]"
    @click="!disabled && (dialogVisible = true)"
  >
    <div class="gva-choose-user-trigger__tags">
      <template v-if="displayRows.length">
        <el-tag
          v-for="item in displayRows"
          :key="item.id"
          size="small"
          :disable-transitions="true"
          :closable="!disabled"
          @close.stop="removeOne(item.id)"
        >
          {{ item.label }}
        </el-tag>
      </template>
      <span v-else class="gva-choose-user-trigger__placeholder">{{ placeholder }}</span>
    </div>
    <el-icon v-if="clearable && displayRows.length && !disabled" class="gva-choose-user-trigger__clear" @click.stop="onClear">
      <circle-close />
    </el-icon>
    <el-icon class="gva-choose-user-trigger__arrow"><arrow-down /></el-icon>
  </div>

  <GvaChooseUserDialog
    v-if="mode === 'dialog'"
    :visible="dialogVisible"
    :model-value="modelValue"
    :multiple="multiple"
    :params="params"
    @update:visible="dialogVisible = $event"
    @update:model-value="(v) => emit('update:modelValue', v)"
    @change="(v, rows) => emit('change', v, rows)"
  />
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import { ArrowDown, CircleClose } from '@element-plus/icons-vue'
  import { ensureUsers, filterUsers, state } from './cache'
  import GvaChooseUserDialog from './chooseUserDialog.vue'

  defineOptions({
    name: 'GvaChooseUser'
  })

  const props = defineProps({
    // 单选 number / 多选 number[]（切换 multiple 由调用方自行清值）
    modelValue: { type: [Number, Array], default: undefined },
    multiple: { type: Boolean, default: false },
    // select：下拉；dialog：输入框触发 + 弹窗（部门/岗位/关键字三维浏览）
    mode: { type: String, default: 'select' },
    // select 形态：候选在缓存内预过滤；dialog 形态：作为初始过滤条件
    params: { type: Object, default: () => ({}) },
    disabled: { type: Boolean, default: false },
    placeholder: { type: String, default: '请选择用户' },
    clearable: { type: Boolean, default: true },
    size: { type: String, default: undefined }
  })

  const emit = defineEmits(['update:modelValue', 'change'])

  const dialogVisible = ref(false)

  // 候选 = 缓存内按 params 预过滤（候选不含冻结用户，后端已过滤）
  const options = computed(() =>
    filterUsers({
      departmentId: props.params && props.params.departmentId,
      postIds: props.params && props.params.postId ? [props.params.postId] : undefined
    })
  )

  // 回显时序：缓存未就绪的未知 ID 先显示 ID 本身，reactive 缓存到位后自动更正
  const displayRows = computed(() =>
    toIds(props.modelValue).map((id) => {
      const u = state.userMap[id]
      return { id, label: u ? u.nickName || u.userName : String(id) }
    })
  )

  const onChange = (value) => {
    const rows = toIds(value).map((id) => state.userMap[id]).filter(Boolean)
    emit('change', value, rows)
  }

  const removeOne = (id) => {
    const rest = toIds(props.modelValue).filter((v) => Number(v) !== Number(id))
    const value = props.multiple ? rest : undefined
    emit('update:modelValue', value)
    emit('change', value, rest.map((v) => state.userMap[v]).filter(Boolean))
  }

  const onClear = () => {
    const value = props.multiple ? [] : undefined
    emit('update:modelValue', value)
    emit('change', value, [])
  }

  function formatUser(u) {
    return u.nickName ? `${u.nickName}（${u.userName}）` : u.userName
  }

  function toIds(value) {
    if (Array.isArray(value)) return value
    if (value === null || value === undefined || value === '') return []
    return [value]
  }

  onMounted(() => {
    ensureUsers()
  })

  defineExpose({
    refresh: () => ensureUsers(true)
  })
</script>

<style scoped lang="scss">
  // dialog 形态触发器：视觉对齐 el-input（tags 版）
  .gva-choose-user-trigger {
    position: relative;
    display: flex;
    align-items: center;
    width: 100%;
    min-height: 32px;
    padding: 1px 30px 1px 8px;
    overflow: hidden;
    cursor: pointer;
    background-color: var(--el-fill-color-blank);
    border: 1px solid var(--el-border-color);
    border-radius: var(--el-border-radius-base);
    transition: border-color 0.2s;

    &:hover {
      border-color: var(--el-border-color-hover);

      .gva-choose-user-trigger__clear {
        opacity: 1;
      }
    }

    &.is-small {
      min-height: 24px;
    }

    &.is-large {
      min-height: 40px;
    }

    &.is-disabled {
      cursor: not-allowed;
      background-color: var(--el-fill-color-light);
      border-color: var(--el-border-color-lighter);
    }
  }

  .gva-choose-user-trigger__tags {
    display: flex;
    flex: 1;
    flex-wrap: wrap;
    gap: 2px 6px;
    max-height: 96px;
    padding: 1px 0;
    overflow-y: auto;
    line-height: 22px;
  }

  .gva-choose-user-trigger__placeholder {
    color: var(--el-text-color-placeholder);
    font-size: 14px;
  }

  .gva-choose-user-trigger__clear {
    position: absolute;
    right: 26px;
    color: var(--el-text-color-placeholder);
    cursor: pointer;
    opacity: 0;
    transition: opacity 0.2s;
  }

  .gva-choose-user-trigger__arrow {
    position: absolute;
    right: 8px;
    color: var(--el-text-color-placeholder);
    font-size: 14px;
  }
</style>
