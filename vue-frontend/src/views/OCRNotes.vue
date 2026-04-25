<template>
  <div class="ocr-page">
    <header class="hero">
      <button @click="$router.push('/menu')">? ??</button>
      <div><p>OCR NOTE STUDIO</p><h1>??????????</h1><span>???????????? Markdown?????????</span></div>
      <el-upload drag multiple :auto-upload="false" :show-file-list="false" accept="image/*,.pdf" :on-change="addFile">
        <div class="drop"><b>???? / PDF</b><small>????</small></div>
      </el-upload>
    </header>

    <main class="grid">
      <section class="card queue">
        <div class="title"><b>????</b><small>{{ activeFiles.length }} ?????</small></div>
        <div v-for="item in files" :key="item.id" :class="['task', { active: selected && selected.id === item.id, removed: item.removed }]" @click="selectedId = item.id">
          <div class="task-head"><b>{{ item.name }}</b><el-tag size="small" :type="tagType(item.status)">{{ statusText(item.status) }}</el-tag></div>
          <el-progress :percentage="item.progress" :status="item.status === 'error' ? 'exception' : undefined" />
          <p v-if="item.error" class="error">{{ item.error }}</p>
          <div class="actions">
            <button v-if="!item.removed" @click.stop="item.removed = true">???</button>
            <button v-else @click.stop="item.removed = false">??</button>
            <button v-if="item.status === 'error'" @click.stop="startOcr(item)">??</button>
          </div>
        </div>
        <p v-if="files.length === 0" class="empty">????????</p>
      </section>

      <section class="card editor">
        <div class="title"><b>Markdown ?? / ??</b><small v-if="selected">{{ selected.name }}</small></div>
        <template v-if="selected">
          <div class="actions toolbar">
            <button :class="{ on: selected.mode === 'source' }" @click="selected.mode = 'source'">??</button>
            <button :class="{ on: selected.mode === 'preview' }" @click="selected.mode = 'preview'">??</button>
            <button :class="{ on: selected.mode === 'split' }" @click="selected.mode = 'split'">??</button>
            <button @click="copy(selected.markdown)">??</button>
            <button @click="download(selected.markdown, selected.name + '.md')">??</button>
          </div>
          <div :class="['md-box', selected.mode]">
            <textarea v-if="selected.mode !== 'preview'" v-model="selected.markdown" placeholder="OCR ????? Markdown???????"></textarea>
            <article v-if="selected.mode !== 'source'" v-html="render(selected.markdown)"></article>
          </div>
        </template>
        <p v-else class="empty big">???????????</p>
      </section>
    </main>

    <section class="bottom">
      <section class="card summary">
        <div class="title"><b>????? Markdown ??</b><small v-if="summary.documentId">??? {{ summary.documentId }}</small></div>
        <div class="actions">
          <el-input v-model="summary.title" placeholder="??????" />
          <button :disabled="!canSummarize || summary.loading" @click="summarize">{{ summary.loading ? '???...' : '???????' }}</button>
          <button :disabled="!summary.markdown" @click="copy(summary.markdown)">??</button>
          <button :disabled="!summary.markdown" @click="download(summary.markdown, (summary.title || 'ocr-note') + '.md')">??</button>
        </div>
        <p v-if="summary.error" class="error">{{ summary.error }}</p>
        <article class="summary-view" v-html="render(summary.markdown || '???????????')"></article>
      </section>

      <section class="card kb">
        <div class="title"><b>?????</b><button @click="loadDocs">??</button></div>
        <div v-for="doc in kb.items" :key="doc.id" class="doc" @click="openDoc(doc)"><b>{{ doc.title || doc.file_name || doc.id }}</b><span>{{ sourceName(doc.source) }} ? {{ doc.created_at }}</span></div>
        <p v-if="kb.items.length === 0" class="empty">?????</p>
        <el-pagination v-model:current-page="kb.page" v-model:page-size="kb.pageSize" :total="kb.total" :page-sizes="[5,10,20,50]" layout="total, sizes, prev, pager, next" @current-change="loadDocs" @size-change="loadDocs" />
      </section>
    </section>

    <el-drawer v-model="kb.drawer" size="52%" title="?????">
      <template v-if="kb.selected"><h2>{{ kb.selected.title || kb.selected.file_name }}</h2><p class="muted">{{ sourceName(kb.selected.source) }}</p><div class="actions"><button @click="copy(kb.selected.content || '')">??</button><button @click="download(kb.selected.content || '', (kb.selected.title || 'document') + '.md')">??</button></div><article class="summary-view" v-html="render(kb.selected.content || '')"></article></template>
    </el-drawer>
  </div>
</template>

<script>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

export default {
  name: 'OCRNotes',
  setup() {
    const files = ref([])
    const selectedId = ref('')
    const summary = reactive({ title: '', markdown: '', documentId: '', loading: false, error: '' })
    const kb = reactive({ items: [], page: 1, pageSize: 10, total: 0, drawer: false, selected: null })
    const activeFiles = computed(() => files.value.filter(i => !i.removed))
    const selected = computed(() => files.value.find(i => i.id === selectedId.value) || null)
    const canSummarize = computed(() => activeFiles.value.some(i => i.status === 'done' && i.markdown.trim()))

    const addFile = upload => {
      const raw = upload.raw
      if (!raw) return
      const item = { id: Date.now() + '-' + Math.random().toString(16).slice(2), file: raw, name: raw.name, taskId: '', documentId: '', status: 'queued', progress: 0, error: '', markdown: '', original: '', mode: 'split', removed: false, es: null, timer: null }
      files.value.push(item)
      if (!selectedId.value) selectedId.value = item.id
      startOcr(item)
    }

    const startOcr = async item => {
      cleanup(item); item.status = 'uploading'; item.progress = 10; item.error = ''
      const fd = new FormData(); fd.append('file', item.file)
      try {
        const res = await api.post('/file/ocr/upload', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
        if (!res.data || res.data.status_code !== 1000) throw new Error(res.data?.status_msg || '?? OCR ????')
        item.taskId = res.data.task_id; item.status = 'processing'; item.progress = 35; watchTask(item)
      } catch (e) { item.status = 'error'; item.progress = 100; item.error = e.message }
    }

    const watchTask = item => {
      const token = localStorage.getItem('token')
      if (!token || !window.EventSource) return poll(item)
      item.es = new EventSource(`/api/file/ocr/tasks/${item.taskId}/events?token=${encodeURIComponent(token)}`)
      item.es.addEventListener('ocr_task', e => updateTask(item, JSON.parse(e.data)))
      item.es.onerror = () => { closeES(item); if (!['done', 'error'].includes(item.status)) poll(item) }
    }
    const poll = item => { item.timer = window.setInterval(() => refresh(item), 2000); refresh(item) }
    const refresh = async item => { const res = await api.get(`/file/ocr/tasks/${item.taskId}`); if (res.data?.task) updateTask(item, res.data.task) }
    const updateTask = (item, task) => {
      if (task.status === 'running') { item.status = 'processing'; item.progress = Math.max(item.progress, 60) }
      if (task.status === 'succeeded') { item.status = 'done'; item.progress = 100; item.documentId = task.document_id || ''; item.markdown = task.result || item.markdown; item.original = item.original || item.markdown; cleanup(item) }
      if (task.status === 'failed') { item.status = 'error'; item.progress = 100; item.error = task.error_msg || 'OCR ??'; cleanup(item) }
    }

    const summarize = async () => {
      const notes = activeFiles.value.filter(i => i.status === 'done' && i.markdown.trim()).map(i => ({ fileName: i.name, markdown: i.markdown, edited: i.markdown !== i.original }))
      summary.loading = true; summary.error = ''
      try {
        const res = await api.post('/file/ocr/notes/summarize', { title: summary.title, notes })
        if (!res.data || res.data.status_code !== 1000) throw new Error(res.data?.status_msg || '????')
        summary.title = res.data.title || summary.title; summary.markdown = res.data.markdown || ''; summary.documentId = res.data.document_id || ''; ElMessage.success('???????????'); loadDocs()
      } catch (e) { summary.error = e.message; ElMessage.error(e.message) } finally { summary.loading = false }
    }
    const loadDocs = async () => { const res = await api.get('/file/documents', { params: { page: kb.page, page_size: kb.pageSize } }); if (res.data?.status_code === 1000) { kb.items = res.data.documents || []; kb.total = res.data.total || 0 } }
    const openDoc = async doc => { kb.drawer = true; kb.selected = { ...doc, content: '???...' }; const res = await api.get(`/file/documents/${doc.id}`); if (res.data?.document) kb.selected = res.data.document }

    const closeES = i => { if (i.es) i.es.close(); i.es = null }
    const cleanup = i => { closeES(i); if (i.timer) clearInterval(i.timer); i.timer = null }
    const copy = async text => { await navigator.clipboard.writeText(text || ''); ElMessage.success('???') }
    const download = (text, name) => { const a = document.createElement('a'); const u = URL.createObjectURL(new Blob([text || ''], { type: 'text/markdown;charset=utf-8' })); a.href = u; a.download = name.replace(/[\\/:*?"<>|]/g, '_'); a.click(); URL.revokeObjectURL(u) }
    const esc = t => String(t || '').replace(/[&<>"]/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c]))
    const render = t => esc(t).replace(/^### (.*)$/gm, '<h3>$1</h3>').replace(/^## (.*)$/gm, '<h2>$1</h2>').replace(/^# (.*)$/gm, '<h1>$1</h1>').replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>').replace(/`([^`]+)`/g, '<code>$1</code>').replace(/\n/g, '<br>')
    const statusText = s => ({ queued: '???', uploading: '???', processing: '???', done: '???', error: '??' }[s] || s)
    const tagType = s => ({ queued: 'info', uploading: 'warning', processing: 'primary', done: 'success', error: 'danger' }[s] || 'info')
    const sourceName = s => ({ upload: '????', ocr: 'OCR ??', ocr_summary: 'OCR ??' }[s] || s || '??')
    onMounted(loadDocs); onBeforeUnmount(() => files.value.forEach(cleanup))
    return { files, selectedId, selected, activeFiles, summary, kb, canSummarize, addFile, startOcr, summarize, loadDocs, openDoc, copy, download, render, statusText, tagType, sourceName }
  }
}
</script>

<style scoped>
.ocr-page{min-height:100vh;padding:26px;background:linear-gradient(135deg,#17324d,#2f5d62 48%,#f2cc8f);color:#18242f}.hero,.card{background:rgba(255,255,255,.93);border-radius:24px;box-shadow:0 24px 70px rgba(12,27,39,.22);padding:20px}.hero{display:grid;grid-template-columns:auto 1fr 340px;gap:20px;align-items:center;margin-bottom:18px}.hero p{letter-spacing:.22em;color:#d98146;font-weight:800}.hero h1{font-size:42px;margin:6px 0}.drop{display:grid;gap:6px}.grid{display:grid;grid-template-columns:360px 1fr;gap:18px}.bottom{display:grid;grid-template-columns:1.1fr .9fr;gap:18px;margin-top:18px}.title,.task-head,.actions,.doc{display:flex;align-items:center;justify-content:space-between;gap:10px}.title{margin-bottom:14px}.title small,.muted{color:#71808b}.task,.doc{display:block;border:1px solid #e4ece7;border-radius:16px;padding:12px;margin-bottom:10px;background:#fbfdf9;cursor:pointer}.task.active{border-color:#d98146;box-shadow:0 12px 30px rgba(217,129,70,.18)}.task.removed{opacity:.45}.actions{justify-content:flex-start;flex-wrap:wrap;margin:10px 0}button{border:0;border-radius:999px;padding:9px 14px;cursor:pointer;background:#e8efe9;color:#17324d;font-weight:700}button:disabled{opacity:.45;cursor:not-allowed}.toolbar .on{background:#17324d;color:white}.md-box{display:grid;gap:12px;min-height:560px}.md-box.split{grid-template-columns:1fr 1fr}textarea,article.summary-view,.md-box article{width:100%;min-height:560px;border:1px solid #e1e8e2;border-radius:18px;padding:16px;background:#fffdf7;box-sizing:border-box;overflow:auto;line-height:1.7}textarea{resize:vertical;font-family:Consolas,monospace}.summary-view{min-height:320px}.error{color:#c0392b}.empty{text-align:center;color:#71808b;padding:28px;border:1px dashed #cbd7d0;border-radius:18px}.big{min-height:520px;display:grid;place-items:center}.doc{display:flex}.doc span{color:#71808b;font-size:13px}@media(max-width:1100px){.hero,.grid,.bottom{grid-template-columns:1fr}.md-box.split{grid-template-columns:1fr}}
</style>
