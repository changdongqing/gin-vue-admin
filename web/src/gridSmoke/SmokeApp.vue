<template>
  <div class="smoke-viewport">
    <div class="gva-container2 overflow-auto">
      <div class="gva-body-h bg-gray-50">
        <SmokeGrid />
        <div class="smoke-footer">BottomInfo 模拟（占位 28px）</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import SmokeGrid from './SmokeGrid.vue'
</script>

<style>
html,
body,
#app {
  height: 100%;
  margin: 0;
}
/* 模拟 layout：滚动容器高度 calc(100% - 4.5rem)（tabs 存在时） */
.smoke-viewport {
  height: 100%;
  box-sizing: border-box;
  padding-top: 3rem;
}
.smoke-viewport > .gva-container2 {
  height: calc(100% - 1.5rem);
  overflow: auto;
  padding: 0 8px;
}
.smoke-footer {
  height: 28px;
  line-height: 28px;
  text-align: center;
  color: #94a3b8;
  font-size: 12px;
}
</style>
