<template>
  <div class="markdown-body" v-html="html" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import DOMPurify from 'dompurify'
import { marked } from 'marked'

const props = defineProps<{ source: string }>()

marked.use({ gfm: true, breaks: true })

const html = computed(() => DOMPurify.sanitize(marked.parse(props.source || '') as string))
</script>
