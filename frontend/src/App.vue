<script setup lang="ts">
</script>

<template>
  <div class="topmost-container">
    <router-view></router-view>
  </div>
</template>


<style scoped>
.topmost-container {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px;
  color: #333;
  background-color: #f9f9f9;
  min-height: 100vh;
  box-sizing: border-box;
}
</style>
