<template>
  <router-view />
</template>

<script setup lang="ts">
import { useUiStore } from './stores/ui'

useUiStore().init()
</script>
