<template>
  <div ref="wrapRef" class="s2-sheet-wrap">
    <SheetComponent
      v-if="sheetProps"
      ref="sheetRef"
      :sheetType="sheetType"
      :dataCfg="sheetProps.dataCfg"
      :options="sheetProps.options"
      :themeCfg="sheetProps.themeCfg"
      :loading="loading"
    />
    <div v-else class="s2-placeholder flex items-center justify-center">
      <slot name="empty">
        <el-empty description="配置行头/列头/数值后开始分析" :image-size="80" />
      </slot>
    </div>
  </div>
</template>

<script setup>
  // @antv/s2-vue SheetComponent 封装：props 响应式驱动（watch 内部自动 setDataCfg/setOptions/render）
  // ——「配置即渲染」零额外代码；sheetType 切换组件内部销毁重建。
  // 实测要点：sheetType 必须经 attrs 显式传入（s2-vue 从 ctx.attrs 取）；宽高由容器测量注入（含 ResizeObserver 自适应）
  import { computed, shallowRef, ref, onMounted, onBeforeUnmount } from 'vue'
  import { SheetComponent } from '@antv/s2-vue'
  import '@antv/s2-vue/dist/s2-vue.min.css'

  const props = defineProps({
    sheetType: { type: String, default: 'pivot' },
    dataCfg: { type: Object, default: null },
    options: { type: Object, default: null },
    themeCfg: { type: Object, default: null },
    loading: { type: Boolean, default: false }
  })

  const sheetRef = shallowRef(null)
  const wrapRef = ref(null)
  const wrapSize = ref({ width: 0, height: 0 })
  let resizeObserver = null

  onMounted(() => {
    const el = wrapRef.value
    if (!el) return
    wrapSize.value = { width: el.offsetWidth, height: el.offsetHeight }
    resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width, height } = entry.contentRect
        if (width > 0 && height > 0) {
          wrapSize.value = { width: Math.floor(width), height: Math.floor(height) }
        }
      }
    })
    resizeObserver.observe(el)
  })

  onBeforeUnmount(() => {
    resizeObserver?.disconnect()
    resizeObserver = null
  })

  const sheetProps = computed(() => {
    if (!props.dataCfg || !props.options) return null
    const options = {
      ...props.options,
      width: wrapSize.value.width || undefined,
      height: wrapSize.value.height || undefined
    }
    return { dataCfg: props.dataCfg, options, themeCfg: props.themeCfg }
  })
</script>

<style lang="scss" scoped>
  .s2-sheet-wrap {
    width: 100%;
    height: 100%;
    overflow: hidden;

    :deep(.antv-s2) {
      height: 100%;
    }
  }

  .s2-placeholder {
    height: 100%;
  }
</style>
