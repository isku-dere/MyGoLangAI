<template>
  <div class="ocr-page">
    <header class="hero">
      <button @click="$router.push('/menu')">{{ t.back }}</button>
      <div>
        <p>OCR NOTE STUDIO</p>
        <h1>{{ t.heroTitle }}</h1>
        <span>{{ t.heroDesc }}</span>
      </div>
      <div class="upload-panel">
        <el-upload drag multiple :auto-upload="false" :show-file-list="false" accept="image/*,.pdf" :on-change="addFile">
          <div class="drop"><b>{{ t.uploadTitle }}</b><small>{{ t.uploadHint }}</small></div>
        </el-upload>
        <button class="primary" :disabled="!canStartOcr" @click="startAllOcr">{{ t.startAll }}</button>
      </div>
    </header>

    <main class="workspace">
      <section class="card queue">
        <div class="title"><b>{{ t.queue }}</b><small>{{ pendingCount }} {{ t.pendingCount }}</small></div>
        <div v-for="item in files" :key="item.id" :class="['task', { active: selected && selected.id === item.id, removed: item.removed }]" @click="selectedId = item.id">
          <div class="task-head"><b>{{ item.name }}</b><el-tag size="small" :type="tagType(item.status)">{{ statusText(item.status) }}</el-tag></div>
          <p v-if="item.elapsedText" class="meta">{{ t.elapsed }} {{ item.elapsedText }}</p>
          <p v-if="item.error" class="error">{{ item.error }}</p>
          <div class="actions">
            <button v-if="!item.removed" @click.stop="item.removed = true">{{ t.exclude }}</button>
            <button v-else @click.stop="item.removed = false">{{ t.restore }}</button>
            <button v-if="item.status === 'error'" @click.stop="startOcr(item)">{{ t.retry }}</button>
          </div>
        </div>
        <p v-if="files.length === 0" class="empty">{{ t.emptyQueue }}</p>
      </section>

      <section class="card preview-card">
        <div class="title"><b>{{ t.previewPanel }}</b><small v-if="selected">{{ selected.name }}</small></div>
        <template v-if="selected">
          <div class="preview-stage">
            <img v-if="selected.previewUrl" :src="selected.previewUrl" :style="previewStyle(selected)" :alt="selected.name" />
            <div v-else class="file-preview"><b>PDF</b><span>{{ selected.name }}</span></div>
          </div>
          <div class="actions preview-actions">
            <button :disabled="!selected.previewUrl || isBusy(selected.status)" @click="rotateSelected(-90)">{{ t.rotateLeft }}</button>
            <button :disabled="!selected.previewUrl || isBusy(selected.status)" @click="rotateSelected(90)">{{ t.rotateRight }}</button>
            <button :disabled="!selected.previewUrl || isBusy(selected.status)" @click="resetRotation">{{ t.reset }}</button>
          </div>
          <p class="muted">{{ selected.previewUrl ? t.rotationTip : t.pdfTip }}</p>
        </template>
        <p v-else class="empty big">{{ t.previewEmpty }}</p>
      </section>

      <section class="card editor">
        <div class="title"><b>{{ t.editorTitle }}</b><small v-if="selected">{{ selected.name }}</small></div>
        <template v-if="selected">
          <div class="actions toolbar">
            <button :class="{ on: selected.mode === 'source' }" @click="selected.mode = 'source'">{{ t.source }}</button>
            <button :class="{ on: selected.mode === 'preview' }" @click="selected.mode = 'preview'">{{ t.preview }}</button>
            <button :class="{ on: selected.mode === 'split' }" @click="selected.mode = 'split'">{{ t.split }}</button>
            <button @click="copy(selected.markdown)">{{ t.copy }}</button>
            <button @click="download(selected.markdown, selected.name + '.md')">{{ t.download }}</button>
          </div>
          <div :class="['md-box', selected.mode]">
            <textarea v-if="selected.mode !== 'preview'" v-model="selected.markdown" :placeholder="t.editorPlaceholder"></textarea>
            <article v-if="selected.mode !== 'source'" v-html="render(selected.markdown)"></article>
          </div>
        </template>
        <p v-else class="empty big">{{ t.selectPrompt }}</p>
      </section>
    </main>

    <section class="bottom">
      <section class="card summary">
        <div class="title"><b>{{ t.summaryTitle }}</b><small v-if="summary.documentId">{{ t.docId }} {{ summary.documentId }}</small></div>
        <div class="actions">
          <el-input v-model="summary.title" :placeholder="t.summaryPlaceholder" />
          <button :disabled="!canSummarize || summary.loading" @click="summarize">{{ summary.loading ? t.summarizing : t.summarize }}</button>
          <button :disabled="!summary.markdown" @click="copy(summary.markdown)">{{ t.copy }}</button>
          <button :disabled="!summary.markdown" @click="download(summary.markdown, (summary.title || 'ocr-note') + '.md')">{{ t.download }}</button>
        </div>
        <p v-if="summary.error" class="error">{{ summary.error }}</p>
        <article class="summary-view" v-html="render(summary.markdown || t.summaryEmpty)"></article>
      </section>

      <section class="card kb">
        <div class="title"><b>{{ t.knowledgeBase }}</b><button @click="loadDocs">{{ t.refresh }}</button></div>
        <div v-for="doc in kb.items" :key="doc.id" class="doc" @click="openDoc(doc)"><b>{{ doc.title || doc.file_name || doc.id }}</b><span>{{ sourceName(doc.source) }} &middot; {{ doc.created_at }}</span></div>
        <p v-if="kb.items.length === 0" class="empty">{{ t.emptyDocs }}</p>
        <el-pagination v-model:current-page="kb.page" v-model:page-size="kb.pageSize" :total="kb.total" :page-sizes="[5,10,20,50]" layout="total, sizes, prev, pager, next" @current-change="loadDocs" @size-change="loadDocs" />
      </section>
    </section>

    <el-drawer v-model="kb.drawer" size="52%" :title="t.drawerTitle">
      <template v-if="kb.selected">
        <h2>{{ kb.selected.title || kb.selected.file_name }}</h2>
        <p class="muted">{{ sourceName(kb.selected.source) }}</p>
        <div class="actions"><button @click="copy(kb.selected.content || '')">{{ t.copy }}</button><button @click="download(kb.selected.content || '', (kb.selected.title || 'document') + '.md')">{{ t.download }}</button></div>
        <article class="summary-view" v-html="render(kb.selected.content || '')"></article>
      </template>
    </el-drawer>
  </div>
</template>

<script>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

const t = {
  back: '\u2190 \u8fd4\u56de',
  heroTitle: '\u624b\u5199\u7b14\u8bb0\u8bc6\u522b\u5de5\u4f5c\u53f0',
  heroDesc: '\u6279\u91cf\u4e0a\u4f20\u56fe\u7247\u6216 PDF\uff0c\u5148\u9884\u89c8\u548c\u65cb\u8f6c\uff0c\u518d\u5e76\u53d1\u8bc6\u522b\u5e76\u5199\u5165\u77e5\u8bc6\u5e93\u3002',
  uploadTitle: '\u4e0a\u4f20\u56fe\u7247 / PDF',
  uploadHint: '\u9009\u62e9\u540e\u5148\u8fdb\u5165\u961f\u5217\uff0c\u70b9\u51fb\u7edf\u4e00\u4e0a\u4f20\u8bc6\u522b\u540e\u5f00\u59cb',
  startAll: '\u7edf\u4e00\u4e0a\u4f20\u8bc6\u522b',
  queue: '\u8bc6\u522b\u961f\u5217',
  pendingCount: '\u4efd\u5f85\u8bc6\u522b',
  activeCount: '\u4efd\u5c06\u53c2\u4e0e\u603b\u7ed3',
  exclude: '\u6392\u9664',
  restore: '\u6062\u590d',
  retry: '\u91cd\u8bd5',
  elapsed: '\u8017\u65f6',
  emptyQueue: '\u8bf7\u5148\u4e0a\u4f20\u624b\u5199\u7b14\u8bb0\u6216 PDF',
  previewPanel: '\u56fe\u7247\u9884\u89c8\u4e0e\u65cb\u8f6c',
  previewEmpty: '\u9009\u62e9\u4e00\u4efd\u6587\u4ef6\u8fdb\u884c\u9884\u89c8',
  rotateLeft: '\u5411\u5de6\u65cb\u8f6c',
  rotateRight: '\u5411\u53f3\u65cb\u8f6c',
  reset: '\u91cd\u7f6e',
  rotationTip: '\u65cb\u8f6c\u540e\u518d\u70b9\u51fb\u7edf\u4e00\u4e0a\u4f20\u8bc6\u522b\uff0cOCR \u5c06\u4f7f\u7528\u65cb\u8f6c\u540e\u7684\u56fe\u7247\u3002',
  pdfTip: 'PDF \u6682\u4e0d\u652f\u6301\u524d\u7aef\u65cb\u8f6c\uff0c\u4f1a\u6309\u539f\u6587\u4ef6\u8bc6\u522b\u3002',
  editorTitle: 'Markdown \u7f16\u8f91 / \u9884\u89c8',
  source: '\u6e90\u7801',
  preview: '\u9884\u89c8',
  split: '\u5206\u5c4f',
  copy: '\u590d\u5236',
  download: '\u4e0b\u8f7d',
  editorPlaceholder: 'OCR \u7ed3\u679c\u4f1a\u5728\u8fd9\u91cc\u663e\u793a\uff0c\u9ed8\u8ba4\u4f7f\u7528\u5206\u5c4f\u9884\u89c8\u3002',
  selectPrompt: '\u9009\u62e9\u4e00\u4efd\u6587\u4ef6\u67e5\u770b\u8bc6\u522b\u7ed3\u679c',
  summaryTitle: '\u4e00\u952e\u603b\u7ed3\u4e3a Markdown \u7b14\u8bb0',
  docId: '\u6587\u6863 ID',
  summaryPlaceholder: '\u7b14\u8bb0\u6807\u9898\uff08\u53ef\u9009\uff09',
  summarizing: '\u603b\u7ed3\u4e2d...',
  summarize: '\u751f\u6210\u603b\u7ed3\u5e76\u5165\u5e93',
  summaryEmpty: '\u7b49\u5f85 OCR \u7ed3\u679c\u540e\u751f\u6210\u603b\u7ed3',
  knowledgeBase: '\u77e5\u8bc6\u5e93\u5185\u5bb9',
  refresh: '\u5237\u65b0',
  emptyDocs: '\u6682\u65e0\u6587\u6863',
  drawerTitle: '\u77e5\u8bc6\u5e93\u6587\u6863',
  uploadFailed: '\u542f\u52a8 OCR \u4efb\u52a1\u5931\u8d25',
  ocrFailed: 'OCR \u5931\u8d25',
  summarizeFailed: '\u603b\u7ed3\u5931\u8d25',
  summarizeSuccess: '\u5df2\u751f\u6210\u603b\u7ed3\u5e76\u5199\u5165\u77e5\u8bc6\u5e93',
  loading: '\u52a0\u8f7d\u4e2d...',
  copied: '\u5df2\u590d\u5236',
  sourceUpload: '\u624b\u52a8\u4e0a\u4f20',
  sourceOcr: 'OCR \u539f\u6587',
  sourceSummary: 'OCR \u603b\u7ed3',
  unknown: '\u672a\u77e5',
  noPending: '\u6ca1\u6709\u9700\u8981\u8bc6\u522b\u7684\u6587\u4ef6'
}

export default {
  name: 'OCRNotes',
  setup() {
    const files = ref([])
    const selectedId = ref('')
    const summary = reactive({ title: '', markdown: '', documentId: '', loading: false, error: '' })
    const kb = reactive({ items: [], page: 1, pageSize: 10, total: 0, drawer: false, selected: null })
    const activeFiles = computed(() => files.value.filter(i => !i.removed))
    const pendingFiles = computed(() => activeFiles.value.filter(i => ['ready', 'error'].includes(i.status)))
    const pendingCount = computed(() => pendingFiles.value.length)
    const selected = computed(() => files.value.find(i => i.id === selectedId.value) || null)
    const canStartOcr = computed(() => pendingFiles.value.length > 0)
    const canSummarize = computed(() => activeFiles.value.some(i => i.status === 'done' && i.markdown.trim()))

    const addFile = upload => {
      const raw = upload.raw
      if (!raw) return
      const isImage = raw.type && raw.type.startsWith('image/')
      const item = {
        id: Date.now() + '-' + Math.random().toString(16).slice(2),
        file: raw,
        name: raw.name,
        previewUrl: isImage ? URL.createObjectURL(raw) : '',
        rotation: 0,
        taskId: '',
        documentId: '',
        status: 'ready',
        error: '',
        markdown: '',
        original: '',
        mode: 'split',
        removed: false,
        es: null,
        timer: null,
        startedAt: 0,
        finishedAt: 0,
        elapsedText: ''
      }
      files.value.push(item)
      selectedId.value = item.id
    }

    const startAllOcr = async () => {
      const targets = pendingFiles.value.slice()
      if (!targets.length) return ElMessage.info(t.noPending)
      await Promise.allSettled(targets.map(item => startOcr(item)))
    }

    const startOcr = async item => {
      cleanup(item)
      item.status = 'uploading'
      item.error = ''
      item.elapsedText = ''
      item.startedAt = performance.now()
      item.finishedAt = 0
      const fd = new FormData()
      try {
        const uploadFile = await buildUploadFile(item)
        fd.append('file', uploadFile, item.name)
        const res = await api.post('/file/ocr/upload', fd, { headers: { 'Content-Type': 'multipart/form-data' } })
        if (!res.data || res.data.status_code !== 1000) throw new Error(res.data?.status_msg || t.uploadFailed)
        item.taskId = res.data.task_id
        item.status = 'processing'
        watchTask(item)
      } catch (e) {
        item.status = 'error'
        item.finishedAt = performance.now()
        item.elapsedText = formatMs(item.finishedAt - item.startedAt)
        item.error = e.message
      }
    }

    const buildUploadFile = async item => {
      if (!item.previewUrl || normalizeRotation(item.rotation) === 0) return item.file
      const blob = await rotateImage(item.previewUrl, item.rotation, item.file.type || 'image/png')
      return new File([blob], item.name, { type: blob.type || item.file.type || 'image/png' })
    }

    const rotateImage = (src, degrees, type) => new Promise((resolve, reject) => {
      const img = new Image()
      img.onload = () => {
        const angle = normalizeRotation(degrees)
        const swap = angle === 90 || angle === 270
        const canvas = document.createElement('canvas')
        canvas.width = swap ? img.naturalHeight : img.naturalWidth
        canvas.height = swap ? img.naturalWidth : img.naturalHeight
        const ctx = canvas.getContext('2d')
        ctx.translate(canvas.width / 2, canvas.height / 2)
        ctx.rotate((angle * Math.PI) / 180)
        ctx.drawImage(img, -img.naturalWidth / 2, -img.naturalHeight / 2)
        canvas.toBlob(blob => blob ? resolve(blob) : reject(new Error('canvas toBlob failed')), type, 0.95)
      }
      img.onerror = reject
      img.src = src
    })

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
      if (task.status === 'running') item.status = 'processing'
      if (task.status === 'succeeded') {
        item.status = 'done'
        item.finishedAt = performance.now()
        item.documentId = task.document_id || ''
        item.markdown = task.result || item.markdown
        item.original = item.original || item.markdown
        item.elapsedText = taskDuration(task) || formatMs(item.finishedAt - item.startedAt)
        cleanup(item)
      }
      if (task.status === 'failed') {
        item.status = 'error'
        item.finishedAt = performance.now()
        item.error = task.error_msg || t.ocrFailed
        item.elapsedText = taskDuration(task) || formatMs(item.finishedAt - item.startedAt)
        cleanup(item)
      }
    }

    const summarize = async () => {
      const notes = activeFiles.value.filter(i => i.status === 'done' && i.markdown.trim()).map(i => ({ fileName: i.name, markdown: i.markdown, edited: i.markdown !== i.original }))
      summary.loading = true; summary.error = ''
      try {
        const res = await api.post('/file/ocr/notes/summarize', { title: summary.title, notes })
        if (!res.data || res.data.status_code !== 1000) throw new Error(res.data?.status_msg || t.summarizeFailed)
        summary.title = res.data.title || summary.title; summary.markdown = res.data.markdown || ''; summary.documentId = res.data.document_id || ''; ElMessage.success(t.summarizeSuccess); loadDocs()
      } catch (e) { summary.error = e.message; ElMessage.error(e.message) } finally { summary.loading = false }
    }
    const loadDocs = async () => { const res = await api.get('/file/documents', { params: { page: kb.page, page_size: kb.pageSize } }); if (res.data?.status_code === 1000) { kb.items = res.data.documents || []; kb.total = res.data.total || 0 } }
    const openDoc = async doc => { kb.drawer = true; kb.selected = { ...doc, content: t.loading }; const res = await api.get(`/file/documents/${doc.id}`); if (res.data?.document) kb.selected = res.data.document }

    const closeES = i => { if (i.es) i.es.close(); i.es = null }
    const cleanup = i => { closeES(i); if (i.timer) clearInterval(i.timer); i.timer = null }
    const releaseItem = i => { cleanup(i); if (i.previewUrl) URL.revokeObjectURL(i.previewUrl) }
    const rotateSelected = delta => { if (selected.value && selected.value.previewUrl) selected.value.rotation = normalizeRotation(selected.value.rotation + delta) }
    const resetRotation = () => { if (selected.value) selected.value.rotation = 0 }
    const previewStyle = item => ({ transform: `rotate(${item.rotation}deg)` })
    const isBusy = status => ['uploading', 'processing'].includes(status)
    const normalizeRotation = value => ((value % 360) + 360) % 360
    const formatMs = ms => {
      if (!Number.isFinite(ms) || ms < 0) return ''
      const seconds = ms / 1000
      return seconds < 60 ? `${seconds.toFixed(1)}s` : `${Math.floor(seconds / 60)}m ${Math.round(seconds % 60)}s`
    }
    const taskDuration = task => {
      const start = Date.parse(task.created_at || '')
      const end = Date.parse(task.updated_at || '')
      return Number.isFinite(start) && Number.isFinite(end) && end >= start ? formatMs(end - start) : ''
    }
    const copy = async text => { await navigator.clipboard.writeText(text || ''); ElMessage.success(t.copied) }
    const download = (text, name) => { const a = document.createElement('a'); const u = URL.createObjectURL(new Blob([text || ''], { type: 'text/markdown;charset=utf-8' })); a.href = u; a.download = name.replace(/[\\/:*?"<>|]/g, '_'); a.click(); URL.revokeObjectURL(u) }
    const esc = value => String(value || '').replace(/[&<>"]/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c]))
    const render = value => esc(value).replace(/^### (.*)$/gm, '<h3>$1</h3>').replace(/^## (.*)$/gm, '<h2>$1</h2>').replace(/^# (.*)$/gm, '<h1>$1</h1>').replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>').replace(/`([^`]+)`/g, '<code>$1</code>').replace(/\n/g, '<br>')
    const statusText = s => ({ ready: '\u5f85\u8bc6\u522b', uploading: '\u4e0a\u4f20\u4e2d', processing: '\u8bc6\u522b\u4e2d', done: '\u5df2\u5b8c\u6210', error: '\u5931\u8d25' }[s] || s)
    const tagType = s => ({ ready: 'info', uploading: 'warning', processing: 'primary', done: 'success', error: 'danger' }[s] || 'info')
    const sourceName = s => ({ upload: t.sourceUpload, ocr: t.sourceOcr, ocr_summary: t.sourceSummary }[s] || s || t.unknown)
    onMounted(loadDocs)
    onBeforeUnmount(() => files.value.forEach(releaseItem))
    return { t, files, selectedId, selected, activeFiles, pendingCount, summary, kb, canStartOcr, canSummarize, addFile, startAllOcr, startOcr, summarize, loadDocs, openDoc, copy, download, render, statusText, tagType, sourceName, rotateSelected, resetRotation, previewStyle, isBusy }
  }
}
</script>

<style scoped>
.ocr-page{min-height:100vh;padding:26px;background:linear-gradient(135deg,#17324d,#2f5d62 48%,#f2cc8f);color:#18242f}.hero,.card{background:rgba(255,255,255,.93);border-radius:24px;box-shadow:0 24px 70px rgba(12,27,39,.22);padding:20px}.hero{display:grid;grid-template-columns:auto 1fr 360px;gap:20px;align-items:center;margin-bottom:18px}.hero p{letter-spacing:.22em;color:#d98146;font-weight:800}.hero h1{font-size:42px;margin:6px 0}.upload-panel{display:grid;gap:12px}.drop{display:grid;gap:6px}.workspace{display:grid;grid-template-columns:300px minmax(280px,.85fr) minmax(440px,1.4fr);gap:18px;align-items:start}.bottom{display:grid;grid-template-columns:1.1fr .9fr;gap:18px;margin-top:18px}.title,.task-head,.actions,.doc{display:flex;align-items:center;justify-content:space-between;gap:10px}.title{margin-bottom:14px}.title small,.muted,.meta{color:#71808b}.task,.doc{display:block;border:1px solid #e4ece7;border-radius:16px;padding:12px;margin-bottom:10px;background:#fbfdf9;cursor:pointer}.task.active{border-color:#d98146;box-shadow:0 12px 30px rgba(217,129,70,.18)}.task.removed{opacity:.45}.actions{justify-content:flex-start;flex-wrap:wrap;margin:10px 0}.primary{background:#d98146;color:white}button{border:0;border-radius:999px;padding:9px 14px;cursor:pointer;background:#e8efe9;color:#17324d;font-weight:700}button:disabled{opacity:.45;cursor:not-allowed}.toolbar .on{background:#17324d;color:white}.preview-stage{height:560px;border:1px solid #e1e8e2;border-radius:18px;background:#fffdf7;display:flex;align-items:center;justify-content:center;overflow:hidden}.preview-stage img{max-width:86%;max-height:86%;transition:transform .2s ease;object-fit:contain;box-shadow:0 12px 32px rgba(23,50,77,.16)}.file-preview{display:grid;gap:10px;place-items:center;color:#71808b}.file-preview b{font-size:52px;color:#d98146}.preview-actions{justify-content:center}.md-box{display:grid;gap:12px;min-height:560px}.md-box.split{grid-template-columns:1fr 1fr}textarea,article.summary-view,.md-box article{width:100%;min-height:560px;border:1px solid #e1e8e2;border-radius:18px;padding:16px;background:#fffdf7;box-sizing:border-box;overflow:auto;line-height:1.7}textarea{resize:vertical;font-family:Consolas,monospace}.summary-view{min-height:320px}.error{color:#c0392b}.empty{text-align:center;color:#71808b;padding:28px;border:1px dashed #cbd7d0;border-radius:18px}.big{min-height:520px;display:grid;place-items:center}.doc{display:flex}.doc span{color:#71808b;font-size:13px}@media(max-width:1280px){.workspace{grid-template-columns:320px 1fr}.editor{grid-column:1/-1}}@media(max-width:900px){.hero,.workspace,.bottom{grid-template-columns:1fr}.md-box.split{grid-template-columns:1fr}.hero h1{font-size:32px}}
</style>
