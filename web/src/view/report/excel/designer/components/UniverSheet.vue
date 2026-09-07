<template>
  <div ref="containerRef" class="univer-sheet"></div>
</template>

<script setup>
  // Univer 封装（0.25.x 实装细节）：
  // ① preset 动态 import + 模块级单例缓存（避免几 MB 进主 bundle）
  // ② zh-CN locale 补丁（0.25.x 缺失 key）
  // ③ sheets:{disableForceStringAlert, disableForceStringMark} 关闭"强制文本数字"告警
  // ④ 快照仅初始化输入；改值走 setCellValue；快照更换由父组件 :key 重建（组件内不 watch）
  // ⑤ 选区监听走 onCommandExecuted（小写 selection 系列命令 id），disposable 卸载时释放
  import { ref, shallowRef, onMounted, onBeforeUnmount } from 'vue'

  const props = defineProps({
    snapshot: { type: Object, default: null },
    readonly: { type: Boolean, default: false }
  })
  const emit = defineEmits(['cell-select'])

  const containerRef = ref(null)
  const univerInstance = shallowRef(null)
  const workbookInstance = shallowRef(null)
  let selectionDisposable = null

  // 模块级单例缓存：createUniver API + locale
  let cachedApi = null
  let cachedLocale = null

  const loadUniver = async () => {
    if (cachedApi) return cachedApi
    const [presetsMod, corePresetMod, zhCNMod] = await Promise.all([
      import('@univerjs/presets'),
      import('@univerjs/preset-sheets-core'),
      import('@univerjs/preset-sheets-core/locales/zh-CN')
    ])
    await import('@univerjs/preset-sheets-core/lib/index.css')
    const zhCN = zhCNMod.default
    // 0.25.x 语言包缺失 key 补丁（缺失时悬停告警显示 key 原文）
    const locales = { zhCN }
    cachedLocale = {
      locale: presetsMod.Locales?.ZH_CN || zhCN,
      locales
    }
    cachedApi = {
      createUniver: presetsMod.createUniver,
      defaultTheme: presetsMod.defaultTheme,
      LocaleType: presetsMod.LocaleType,
      mergeLocales: presetsMod.mergeLocales,
      UniverSheetsCorePreset: corePresetMod.UniverSheetsCorePreset
    }
    return cachedApi
  }

  const createEmptyWorkbook = () => ({
    id: 'empty-workbook',
    sheets: {
      sheet1: {
        id: 'sheet1',
        name: 'Sheet1',
        rowCount: 40,
        columnCount: 20,
        cellData: {}
      }
    },
    views: [{}]
  })

  const emitSelection = (workbook) => {
    try {
      const sheet = workbook.getActiveSheet()
      if (!sheet) return
      const sel = sheet.getSelection()?.getCurrent()
      if (!sel) return
      const row = sel.actualRow ?? sel.row
      const col = sel.actualColumn ?? sel.column
      const cell = sheet.getCell(row, col)
      emit('cell-select', {
        sheetName: sheet.getName(),
        row,
        col,
        value: cell?.getValue?.() ?? ''
      })
    } catch (e) {
      /* 选区解析失败忽略 */
    }
  }

  onMounted(async () => {
    const api = await loadUniver()
    const { createUniver, defaultTheme, LocaleType, mergeLocales, UniverSheetsCorePreset } = api
    const zhCN = (await import('@univerjs/preset-sheets-core/locales/zh-CN')).default
    const { univerAPI, univer } = createUniver({
      locale: LocaleType.ZH_CN,
      locales: {
        [LocaleType.ZH_CN]: mergeLocales(zhCN, {
          'sheets-ui': {
            info: {
              error: '错误',
              forceStringInfo: '以文本形式存储的数字'
            }
          }
        })
      },
      theme: defaultTheme,
      presets: [
        UniverSheetsCorePreset({
          container: containerRef.value,
          sheets: {
            disableForceStringAlert: true,
            disableForceStringMark: true
          }
        })
      ]
    })
    univerInstance.value = univer
    const workbook = univerAPI.createWorkbook(props.snapshot || createEmptyWorkbook())
    workbookInstance.value = workbook
    if (props.readonly) {
      workbook.setEditable(false)
    }
    // 选区监听：小写 selection 系列命令（SetSelections/MoveSelection/SelectAll…）
    selectionDisposable = univerAPI.onCommandExecuted((command) => {
      const id = (command?.id || '').toLowerCase()
      if (id.includes('selection')) {
        emitSelection(workbook)
      }
    })
    emitSelection(workbook)
  })

  onBeforeUnmount(() => {
    // 铁律：disposable → workbook → univer 顺序释放
    if (selectionDisposable) {
      selectionDisposable.dispose()
      selectionDisposable = null
    }
    try {
      workbookInstance.value?.dispose?.()
    } catch (e) {
      /* ignore */
    }
    try {
      univerInstance.value?.dispose?.()
    } catch (e) {
      /* ignore */
    }
  })

  const getActiveWorkbook = () => workbookInstance.value

  const getSnapshot = () => {
    const wb = workbookInstance.value
    return wb ? wb.save() : null
  }

  const setCellValue = (row, col, value) => {
    const wb = workbookInstance.value
    if (!wb) return
    wb.getActiveSheet().getRange(row, col, 1, 1).setValue(value)
  }

  const getSelection = () => {
    const wb = workbookInstance.value
    if (!wb) return null
    const sheet = wb.getActiveSheet()
    const sel = sheet.getSelection()?.getCurrent()
    if (!sel) return null
    return {
      sheetId: sheet.getSheetId(),
      sheetName: sheet.getName(),
      row: sel.actualRow ?? sel.row ?? 0,
      col: sel.actualColumn ?? sel.column ?? 0
    }
  }

  defineExpose({ getSnapshot, setCellValue, getSelection, getActiveWorkbook })
</script>

<style lang="scss" scoped>
  .univer-sheet {
    width: 100%;
    height: 100%;
    overflow: hidden;
  }
</style>
