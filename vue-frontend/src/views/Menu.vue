<template>
  <div class="readme-page">
    <header class="readme-header">
      <div>
        <p class="page-label">PROJECT README</p>
        <h1>GopherAI 项目说明</h1>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="$router.push('/ai-chat')">AI 对话</el-button>
        <el-button @click="$router.push('/ocr-notes')">OCR 笔记</el-button>
        <el-button type="danger" plain @click="handleLogout">退出登录</el-button>
      </div>
    </header>

    <main class="readme-main">
      <section class="entry-grid" aria-label="功能入口">
        <button class="entry-button primary" @click="$router.push('/ai-chat')">
          <strong>AI 对话</strong>
          <span>普通问答、流式响应、多会话管理</span>
        </button>
        <button class="entry-button" @click="$router.push('/ai-chat')">
          <strong>RAG 知识库</strong>
          <span>上传文档，按用户隔离检索私有知识</span>
        </button>
        <button class="entry-button" @click="$router.push('/ocr-notes')">
          <strong>OCR 笔记</strong>
          <span>异步识别手写内容，总结后保存入库</span>
        </button>
      </section>

      <section class="readme-box" aria-label="项目 Markdown 说明">
        <div class="readme-toolbar">
          <span>README.md</span>
          <span>生产环境：Nginx HTTPS + Docker Compose</span>
        </div>
        <div class="markdown-view" v-html="renderedReadme"></div>
      </section>
    </main>
  </div>
</template>

<script>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

const projectReadme = `# GopherAI

一个基于 Go + Gin 的 AI 知识库与 OCR 手写笔记整理系统。

项目定位：
- 面向个人知识管理和简历展示的完整 AI 应用。
- 支持普通 AI 对话、RAG 私有知识库问答、OCR 手写笔记整理。
- 线上通过 Docker Compose 部署，Nginx 提供 HTTPS 反向代理。

## 核心功能

### 1. AI 对话
- 支持多会话管理。
- 支持普通问答和流式响应。
- 支持 Markdown 消息渲染，包括标题、列表、引用、代码块、加粗和行内代码。

### 2. RAG 知识库
- 用户可上传 Markdown / TXT 文档。
- 文档完整内容保存到 MySQL。
- 文档向量写入 Redis Vector。
- 检索时按 username 做用户隔离，只召回当前用户自己的知识库内容。
- 删除知识库文档时同步清理 Redis 向量数据，避免 MySQL 与 Redis 残留不一致。

### 3. OCR 手写笔记整理
- 支持上传图片或 PDF 创建 OCR 任务。
- 上传后立即创建任务并进入 RabbitMQ 队列。
- app 内置 OCR worker 并发消费任务，避免外部 PaddleOCR 接口阻塞上传请求。
- 前端通过任务状态展示 pending / running / succeeded / failed。
- OCR 原文不会直接进入知识库，只有用户手动保存总结后的 Markdown 笔记才会写入 RAG。

### 4. 注册登录
- 支持邮箱验证码注册。
- 注册成功后返回随机账号 username。
- 邮箱唯一性校验，避免重复邮箱注册。

## 技术栈

- 后端：Go、Gin
- 数据库：MySQL
- 向量存储：Redis Vector
- 消息队列：RabbitMQ
- OCR：PaddleOCR 外部 API
- 部署：Docker Compose
- 反向代理：Nginx HTTPS

## 生产服务

- app：后端 API 服务与 OCR worker
- nginx：前端静态资源与 HTTPS 反向代理
- mysql：业务数据、用户、会话、文档、OCR 任务
- redis：RAG 向量索引
- rabbitmq：OCR 异步任务队列

## 数据流

### 普通对话
用户 -> Nginx -> Go API -> 大模型接口 -> 返回回答

### RAG 问答
用户问题 -> Go API -> Redis Vector 检索 -> 拼接上下文 -> 大模型接口 -> 返回回答

### OCR 笔记
上传文件 -> 创建 MySQL 任务 -> 投递 RabbitMQ -> worker 调用 PaddleOCR -> 写入 OCR 结果 -> 用户总结并保存 Markdown -> 写入知识库

## 部署与稳定性

- 公网 API 统一使用 /api 路径。
- Nginx 将 /api 转发到后端 /api/v1。
- MySQL、Redis、RabbitMQ 使用 Docker volume 持久化。
- OCR 任务使用队列异步化，失败即终态，由用户重新上传重试。
- app 重启后会恢复未完成 OCR 任务，避免任务长期卡在处理中。`

const text = {
  confirmLogout: '确定要退出登录吗？',
  warning: '提示',
  confirm: '确定',
  cancel: '取消',
  logoutSuccess: '已退出登录'
}

const escapeHtml = (value) => String(value)
  .replace(/&/g, '&amp;')
  .replace(/</g, '&lt;')
  .replace(/>/g, '&gt;')
  .replace(/"/g, '&quot;')
  .replace(/'/g, '&#39;')

const renderInlineMarkdown = (value) => escapeHtml(value)
  .replace(/`([^`]+)`/g, '<code>$1</code>')
  .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
  .replace(/\*([^*]+)\*/g, '<em>$1</em>')

const flushParagraph = (paragraph, output) => {
  if (paragraph.length > 0) {
    output.push(`<p>${paragraph.map(renderInlineMarkdown).join('<br>')}</p>`)
    paragraph.length = 0
  }
}

const renderMarkdown = (markdown) => {
  const lines = String(markdown).replace(/\r\n/g, '\n').split('\n')
  const output = []
  const paragraph = []
  let listOpen = false

  const closeList = () => {
    if (listOpen) {
      output.push('</ul>')
      listOpen = false
    }
  }

  for (const line of lines) {
    if (!line.trim()) {
      flushParagraph(paragraph, output)
      closeList()
      continue
    }

    const heading = line.match(/^(#{1,3})\s+(.+)$/)
    if (heading) {
      flushParagraph(paragraph, output)
      closeList()
      const level = heading[1].length
      output.push(`<h${level}>${renderInlineMarkdown(heading[2])}</h${level}>`)
      continue
    }

    const unordered = line.match(/^-\s+(.+)$/)
    if (unordered) {
      flushParagraph(paragraph, output)
      if (!listOpen) {
        output.push('<ul>')
        listOpen = true
      }
      output.push(`<li>${renderInlineMarkdown(unordered[1])}</li>`)
      continue
    }

    paragraph.push(line)
  }

  flushParagraph(paragraph, output)
  closeList()
  return output.join('')
}

export default {
  name: 'MenuView',
  setup() {
    const router = useRouter()
    const renderedReadme = computed(() => renderMarkdown(projectReadme))
    const handleLogout = async () => {
      try {
        await ElMessageBox.confirm(text.confirmLogout, text.warning, {
          confirmButtonText: text.confirm,
          cancelButtonText: text.cancel,
          type: 'warning'
        })
        localStorage.removeItem('token')
        ElMessage.success(text.logoutSuccess)
        router.push('/login')
      } catch {
        // User cancelled logout.
      }
    }
    return { renderedReadme, handleLogout }
  }
}
</script>

<style scoped>
.readme-page { min-height: 100vh; background: #f4f6f8; color: #1f2937; }
.readme-header { display: flex; align-items: center; justify-content: space-between; gap: 18px; padding: 24px clamp(16px, 4vw, 48px); background: #fff; border-bottom: 1px solid #e5e7eb; }
.page-label { margin: 0 0 6px; color: #667085; font-size: 12px; font-weight: 700; letter-spacing: 0; }
.readme-header h1 { margin: 0; font-size: 28px; line-height: 1.25; color: #111827; }
.header-actions { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 10px; }
.header-actions .el-button { margin-left: 0; }
.readme-main { width: min(1080px, calc(100% - 32px)); margin: 0 auto; padding: 28px 0 48px; }
.entry-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 16px; margin-bottom: 22px; }
.entry-button { min-height: 132px; padding: 24px; text-align: left; border: 1px solid #d8dee4; border-radius: 8px; background: #fff; color: #1f2937; cursor: pointer; box-shadow: 0 10px 24px rgba(17, 24, 39, 0.06); transition: transform .16s ease, border-color .16s ease, box-shadow .16s ease; }
.entry-button:hover { transform: translateY(-3px); border-color: #7aa89f; box-shadow: 0 16px 30px rgba(17, 24, 39, 0.1); }
.entry-button.primary { background: #172033; color: #fff; border-color: #172033; }
.entry-button strong { display: block; margin-bottom: 12px; font-size: 24px; line-height: 1.2; }
.entry-button span { display: block; color: #667085; font-size: 15px; line-height: 1.6; }
.entry-button.primary span { color: #d8dee9; }
.readme-box { overflow: hidden; border: 1px solid #d8dee4; border-radius: 8px; background: #fff; box-shadow: 0 14px 34px rgba(17, 24, 39, 0.08); }
.readme-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 12px 16px; background: #f6f8fa; border-bottom: 1px solid #d8dee4; color: #57606a; font-size: 13px; }
.readme-toolbar span:first-child { color: #24292f; font-weight: 700; }
.markdown-view { min-height: 620px; max-height: calc(100vh - 360px); margin: 0; padding: 28px 32px; overflow: auto; overflow-wrap: anywhere; word-break: break-word; color: #24292f; font-size: 15px; line-height: 1.78; background: #fff; }
.markdown-view :deep(h1) { margin: 0 0 18px; padding-bottom: 12px; border-bottom: 1px solid #d8dee4; color: #111827; font-size: 32px; line-height: 1.25; }
.markdown-view :deep(h2) { margin: 30px 0 12px; padding-bottom: 8px; border-bottom: 1px solid #edf0f3; color: #111827; font-size: 23px; line-height: 1.35; }
.markdown-view :deep(h3) { margin: 22px 0 8px; color: #1f2937; font-size: 18px; line-height: 1.4; }
.markdown-view :deep(p) { margin: 0 0 14px; }
.markdown-view :deep(ul) { margin: 0 0 16px; padding-left: 24px; }
.markdown-view :deep(li) { margin: 6px 0; }
.markdown-view :deep(code) { padding: 2px 5px; border-radius: 4px; background: #f6f8fa; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace; font-size: .92em; }

@media (max-width: 720px) {
  .readme-header { align-items: flex-start; flex-direction: column; padding: 18px 14px; }
  .readme-header h1 { font-size: 24px; }
  .header-actions { width: 100%; display: grid; grid-template-columns: 1fr 1fr; }
  .header-actions .el-button { min-height: 38px; }
  .header-actions .el-button:last-child { grid-column: 1 / -1; }
  .readme-main { width: calc(100% - 20px); padding: 14px 0 28px; }
  .entry-grid { grid-template-columns: 1fr; gap: 12px; margin-bottom: 14px; }
  .entry-button { min-height: 104px; padding: 18px; }
  .entry-button strong { font-size: 21px; }
  .readme-toolbar { align-items: flex-start; flex-direction: column; padding: 10px 12px; }
  .markdown-view { min-height: 560px; max-height: none; padding: 18px 14px; font-size: 13px; line-height: 1.68; }
  .markdown-view :deep(h1) { font-size: 26px; }
  .markdown-view :deep(h2) { font-size: 20px; }
}
</style>
