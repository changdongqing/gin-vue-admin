<template>
  <div class="config-panel">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="数据" name="data">
        <div class="px-1">
          <!-- 视图 -->
          <div class="config-block">
            <div class="config-label">视图</div>
            <el-radio-group :model-value="config.sheetType" size="small" @update:model-value="emit('set-sheet-type', $event)">
              <el-radio-button value="pivot">透视</el-radio-button>
              <el-radio-button value="table">明细</el-radio-button>
            </el-radio-group>
          </div>

          <!-- 行头 -->
          <div class="config-block" :class="{ disabled: isTable }">
            <div class="config-label">行头（有序）</div>
            <el-empty v-if="!config.fields.rows.length" description="从左栏添加" :image-size="40" />
            <div v-for="(r, idx) in config.fields.rows" :key="r" class="chip-item">
              <span class="flex-1 truncate">{{ r }}</span>
              <el-button icon="top" size="small" link :disabled="idx === 0" @click="move(config.fields.rows, idx, -1)" />
              <el-button icon="bottom" size="small" link :disabled="idx === config.fields.rows.length - 1" @click="move(config.fields.rows, idx, 1)" />
              <el-button icon="close" size="small" link @click="removeAt(config.fields.rows, idx)" />
            </div>
          </div>

          <!-- 列头 -->
          <div class="config-block" :class="{ disabled: isTable }">
            <div class="config-label">列头（有序）</div>
            <el-empty v-if="!config.fields.columns.length" description="从左栏添加" :image-size="40" />
            <div v-for="(c, idx) in config.fields.columns" :key="c" class="chip-item">
              <span class="flex-1 truncate">{{ c }}</span>
              <el-button icon="top" size="small" link :disabled="idx === 0" @click="move(config.fields.columns, idx, -1)" />
              <el-button icon="bottom" size="small" link :disabled="idx === config.fields.columns.length - 1" @click="move(config.fields.columns, idx, 1)" />
              <el-button icon="close" size="small" link @click="removeAt(config.fields.columns, idx)" />
            </div>
          </div>

          <!-- 数值列 -->
          <div class="config-block">
            <div class="config-label">数值列</div>
            <el-empty v-if="!config.fields.values.length" description="从左栏添加度量" :image-size="40" />
            <div v-for="(v, idx) in config.fields.values" :key="`${v.field}__${v.aggregation}__${idx}`" class="value-item">
              <div class="flex items-center gap-2">
                <span class="flex-1 truncate font-bold">{{ v.field }}</span>
                <el-select :model-value="v.aggregation" size="small" style="width: 130px" @update:model-value="setAggregation(idx, $event)">
                  <el-option v-for="o in AGGREGATION_OPTIONS" :key="o.value" :label="o.label" :value="o.value" />
                </el-select>
                <el-button icon="close" size="small" link @click="config.fields.values.splice(idx, 1)" />
              </div>
              <div class="flex items-center gap-2 mt-1">
                <el-input v-model="v.alias" size="small" placeholder="别名" style="width: 130px" />
                <el-input v-model="v.format" size="small" placeholder="格式化，如 #,##0.00" style="width: 150px" />
              </div>
            </div>
          </div>

          <!-- 数值置于 -->
          <div class="config-block" :class="{ disabled: isTable }">
            <div class="config-label">数值置于</div>
            <el-radio-group :model-value="config.fields.valueInRow" size="small" :disabled="isTable" @update:model-value="emit('set-value-in-row', $event)">
              <el-radio-button :value="false">列头</el-radio-button>
              <el-radio-button :value="true">行头</el-radio-button>
            </el-radio-group>
          </div>

          <!-- 合计（透视专属） -->
          <template v-if="!isTable">
            <div class="config-block">
              <div class="config-label">
                行小计/总计
                <el-tooltip content="S2 合计为维度级全局 SUM：SUM 列精确，AVG 列是「平均的平均」，NONE 列也会被求和；精确汇总建议 SQL 预聚合" placement="top">
                  <el-icon class="text-gray-400"><info-filled /></el-icon>
                </el-tooltip>
              </div>
              <div class="flex flex-wrap items-center gap-3">
                <el-switch v-model="config.options.totals.row.showGrandTotals" active-text="总计" />
                <el-switch v-model="config.options.totals.row.showSubTotals" active-text="小计" />
              </div>
              <div v-if="config.options.totals.row.showSubTotals" class="text-xs text-gray-400 mt-1">
                小计维度：{{ subTotalsDimsText || '（无可用维度）' }}
              </div>
            </div>
            <div class="config-block">
              <div class="config-label">列小计/总计</div>
              <div class="flex flex-wrap items-center gap-3">
                <el-switch v-model="config.options.totals.col.showGrandTotals" active-text="总计" />
                <el-switch v-model="config.options.totals.col.showSubTotals" active-text="小计" />
              </div>
            </div>
          </template>
        </div>
      </el-tab-pane>

      <el-tab-pane label="样式" name="style">
        <div class="px-1">
          <div class="config-block">
            <div class="config-label">主题</div>
            <el-select v-model="config.theme.name" size="small">
              <el-option v-for="o in THEME_OPTIONS" :key="o.value" :label="o.label" :value="o.value" />
            </el-select>
          </div>
          <div class="config-block" :class="{ disabled: isTable }">
            <div class="config-label">显示模式</div>
            <el-radio-group v-model="config.options.hierarchyType" size="small" :disabled="isTable">
              <el-radio-button v-for="o in HIERARCHY_OPTIONS" :key="o.value" :value="o.value">{{ o.label }}</el-radio-button>
            </el-radio-group>
          </div>
          <div class="config-block">
            <div class="config-label">宽度调整</div>
            <el-radio-group v-model="config.options.layoutWidthType" size="small">
              <el-radio-button v-for="o in LAYOUT_WIDTH_OPTIONS" :key="o.value" :value="o.value">{{ o.label }}</el-radio-button>
            </el-radio-group>
          </div>
          <div class="config-block" :class="{ disabled: isTable }">
            <div class="config-label">冻结行头</div>
            <el-switch v-model="config.options.frozenRowHeader" :disabled="isTable" />
          </div>
          <div class="config-block">
            <div class="config-label">行序号</div>
            <el-switch v-model="config.options.showSeriesNumber" />
          </div>
          <div class="config-block">
            <div class="config-label">分页</div>
            <div class="flex items-center gap-3">
              <el-switch v-model="config.options.pagination.open" />
              <template v-if="config.options.pagination.open">
                <span class="text-xs text-gray-400">每页</span>
                <el-select v-model="config.options.pagination.pageSize" size="small" style="width: 90px" @update:model-value="emit('reset-page')">
                  <el-option v-for="s in PAGE_SIZE_OPTIONS" :key="s" :label="String(s)" :value="s" />
                </el-select>
              </template>
            </div>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
  // 右栏容器（数据/样式两 Tab）+ 联动规则：
  // 视图=明细 → 禁用行头/列头/数值置于/小计/冻结行头；行小计开启 → subTotalsDimensions=rows 去最外层
  import { ref, computed, watch } from 'vue'
  import { AGGREGATION_OPTIONS, THEME_OPTIONS, HIERARCHY_OPTIONS, LAYOUT_WIDTH_OPTIONS, PAGE_SIZE_OPTIONS } from '@/api/report/analysisReport'
  import { calcSubTotalsDimensions } from '../utils/s2Config'

  const props = defineProps({
    config: { type: Object, required: true }
  })
  const emit = defineEmits(['set-sheet-type', 'set-value-in-row', 'reset-page'])

  const activeTab = ref('data')
  const isTable = computed(() => props.config.sheetType === 'table')

  const subTotalsDimsText = computed(() => calcSubTotalsDimensions(props.config.fields.rows).join(' / '))

  // 行小计开启 → subTotalsDimensions 自动 = rows 去最外层（watch rows 重算）
  watch(
    () => [...props.config.fields.rows],
    (rows) => {
      if (props.config.options.totals.row.showSubTotals) {
        props.config.options.totals.row.subTotalsDimensions = calcSubTotalsDimensions(rows)
      }
    }
  )
  watch(
    () => props.config.options.totals.row.showSubTotals,
    (open) => {
      if (open) {
        props.config.options.totals.row.subTotalsDimensions = calcSubTotalsDimensions(props.config.fields.rows)
      }
    }
  )

  const move = (arr, idx, delta) => {
    const target = idx + delta
    if (target < 0 || target >= arr.length) return
    const [item] = arr.splice(idx, 1)
    arr.splice(target, 0, item)
    emit('reset-page')
  }
  const removeAt = (arr, idx) => {
    arr.splice(idx, 1)
    emit('reset-page')
  }
  const setAggregation = (idx, agg) => {
    props.config.fields.values[idx].aggregation = agg
    emit('reset-page')
  }
</script>

<style lang="scss" scoped>
  .config-panel {
    height: 100%;
    overflow: auto;

    .config-block {
      margin-bottom: 14px;

      &.disabled {
        opacity: 0.45;
      }

      .config-label {
        font-size: 13px;
        font-weight: bold;
        margin-bottom: 6px;
        display: flex;
        align-items: center;
        gap: 4px;
      }
    }

    .chip-item,
    .value-item {
      border: 1px solid var(--el-border-color-lighter);
      border-radius: 4px;
      padding: 4px 8px;
      margin-bottom: 4px;
      font-size: 13px;
      background: var(--el-bg-color);
    }
  }
</style>
