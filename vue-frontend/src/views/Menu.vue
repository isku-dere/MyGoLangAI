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
      <section class="readme-box" aria-label="项目 Markdown 说明">
        <div class="readme-toolbar">
          <span>README.md</span>
          <span>生产环境：Nginx HTTPS + Docker Compose</span>
        </div>
        <pre class="markdown-view">{{ projectReadme }}</pre>
      </section>
    </main>
  </div>
</template>

<script>
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

export default {
  name: 'MenuView',
  setup() {
    const router = useRouter()
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
    return { projectReadme, handleLogout }
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
.readme-box { overflow: hidden; border: 1px solid #d8dee4; border-radius: 8px; background: #fff; box-shadow: 0 14px 34px rgba(17, 24, 39, 0.08); }
.readme-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 12px 16px; background: #f6f8fa; border-bottom: 1px solid #d8dee4; color: #57606a; font-size: 13px; }
.readme-toolbar span:first-child { color: #24292f; font-weight: 700; }
.markdown-view { min-height: 620px; max-height: calc(100vh - 210px); margin: 0; padding: 28px 32px; overflow: auto; white-space: pre-wrap; overflow-wrap: anywhere; word-break: break-word; color: #24292f; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace; font-size: 15px; line-height: 1.78; background: #fff; }

@media (max-width: 720px) {
  .readme-header { align-items: flex-start; flex-direction: column; padding: 18px 14px; }
  .readme-header h1 { font-size: 24px; }
  .header-actions { width: 100%; display: grid; grid-template-columns: 1fr 1fr; }
  .header-actions .el-button { min-height: 38px; }
  .header-actions .el-button:last-child { grid-column: 1 / -1; }
  .readme-main { width: calc(100% - 20px); padding: 14px 0 28px; }
  .readme-toolbar { align-items: flex-start; flex-direction: column; padding: 10px 12px; }
  .markdown-view { min-height: 560px; max-height: none; padding: 18px 14px; font-size: 13px; line-height: 1.68; }
}
</style>
