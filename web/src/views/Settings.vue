<template>
  <div class="settings-page">
    <div class="settings-tabs">
      <router-link to="/settings/config" class="tab-item" :class="{ active: isConfig }">VPN 配置</router-link>
      <router-link to="/settings/system" class="tab-item" :class="{ active: isSystem }">系统配置</router-link>
    </div>
    <div class="settings-content">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const isConfig = computed(() => route.path === '/settings/config' || route.path === '/settings')
const isSystem = computed(() => route.path === '/settings/system')
</script>

<style scoped>
.settings-page {
  display: flex;
  flex-direction: column;
  gap: 0;
}
.settings-tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid var(--border);
  margin: -0.25rem 0 1rem;
  padding: 0 0.25rem 0 0;
}
.tab-item {
  padding: 0.6rem 1.25rem;
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--muted);
  text-decoration: none;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
  transition: color 0.15s, border-color 0.15s;
}
.tab-item:hover {
  color: var(--accent);
}
.tab-item.active {
  color: var(--accent);
  border-bottom-color: var(--accent);
}
.settings-content {
  flex: 1;
}
</style>
