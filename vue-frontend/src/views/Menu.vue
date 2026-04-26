<template>
  <div class="home-shell">
    <header class="top-nav">
      <div class="brand">
        <span class="brand-mark">G</span>
        <div>
          <strong>{{ text.brand }}</strong>
          <span>{{ text.subtitle }}</span>
        </div>
      </div>
      <nav class="nav-actions">
        <el-button plain @click="$router.push('/ai-chat')">{{ text.chatTitle }}</el-button>
        <el-button plain @click="$router.push('/ocr-notes')">{{ text.ocrTitle }}</el-button>
        <el-button type="danger" plain @click="handleLogout">{{ text.logout }}</el-button>
      </nav>
    </header>

    <main class="home-main">
      <section class="hero">
        <div class="hero-copy">
          <p class="eyebrow">{{ text.eyebrow }}</p>
          <h1>{{ text.heroTitle }}</h1>
          <p class="hero-desc">{{ text.heroDesc }}</p>
          <div class="hero-actions">
            <el-button type="primary" size="large" @click="$router.push('/ai-chat')">{{ text.startChat }}</el-button>
            <el-button size="large" @click="$router.push('/ocr-notes')">{{ text.openOcr }}</el-button>
          </div>
        </div>
        <div class="status-panel">
          <div v-for="metric in metrics" :key="metric.label" class="metric">
            <span>{{ metric.label }}</span>
            <strong>{{ metric.value }}</strong>
          </div>
        </div>
      </section>

      <section class="section">
        <div class="section-heading">
          <p>{{ text.featureEyebrow }}</p>
          <h2>{{ text.featureTitle }}</h2>
        </div>
        <div class="feature-grid">
          <article class="feature-card" @click="$router.push('/ai-chat')">
            <el-icon><ChatDotRound /></el-icon>
            <h3>{{ text.chatTitle }}</h3>
            <p>{{ text.chatDesc }}</p>
          </article>
          <article class="feature-card" @click="$router.push('/ai-chat')">
            <el-icon><Search /></el-icon>
            <h3>{{ text.ragTitle }}</h3>
            <p>{{ text.ragDesc }}</p>
          </article>
          <article class="feature-card" @click="$router.push('/ocr-notes')">
            <el-icon><Document /></el-icon>
            <h3>{{ text.ocrTitle }}</h3>
            <p>{{ text.ocrDesc }}</p>
          </article>
        </div>
      </section>

      <section class="section split-section">
        <div>
          <div class="section-heading compact">
            <p>{{ text.archEyebrow }}</p>
            <h2>{{ text.archTitle }}</h2>
          </div>
          <div class="flow-list">
            <div v-for="step in flow" :key="step.title" class="flow-step">
              <span>{{ step.index }}</span>
              <div>
                <h3>{{ step.title }}</h3>
                <p>{{ step.desc }}</p>
              </div>
            </div>
          </div>
        </div>
        <aside class="stack-panel">
          <h2>{{ text.stackTitle }}</h2>
          <div class="stack-list">
            <span v-for="item in stack" :key="item">{{ item }}</span>
          </div>
        </aside>
      </section>
    </main>
  </div>
</template>

<script>
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ChatDotRound, Document, Search } from '@element-plus/icons-vue'

const text = {
  brand: 'GopherAI',
  subtitle: 'AI 知识库与 OCR 笔记系统',
  logout: '退出登录',
  eyebrow: '个人项目工程展示',
  heroTitle: '面向知识问答与手写笔记整理的 AI 应用',
  heroDesc: '系统覆盖账号注册登录、流式 AI 对话、RAG 私有知识库、OCR 异步任务、Markdown 笔记总结与保存，并通过容器化方式部署到 HTTPS 站点。',
  startChat: '进入 AI 对话',
  openOcr: '整理 OCR 笔记',
  featureEyebrow: '核心能力',
  featureTitle: '从文档沉淀到智能问答',
  chatTitle: 'AI 对话',
  chatDesc: '支持普通问答和流式响应，聊天历史按会话管理。',
  ragTitle: 'RAG 知识库',
  ragDesc: '文档正文保存在 MySQL，向量写入 Redis Vector，并按用户隔离检索。',
  ocrTitle: 'OCR 手写笔记整理',
  ocrDesc: '上传图片或 PDF 后进入 RabbitMQ 队列，识别结果可总结为 Markdown 并手动入库。',
  archEyebrow: '架构流程',
  archTitle: '生产环境链路',
  stackTitle: '技术栈',
  confirmLogout: '确定要退出登录吗？',
  warning: '提示',
  confirm: '确定',
  cancel: '取消',
  logoutSuccess: '已退出登录'
}

const metrics = [
  { label: '公网入口', value: 'Nginx HTTPS' },
  { label: '任务队列', value: 'RabbitMQ' },
  { label: '向量检索', value: 'Redis Vector' }
]

const flow = [
  { index: '01', title: '用户请求进入网关', desc: '公网路径统一使用 /api，由 Nginx 转发到后端服务。' },
  { index: '02', title: '业务服务处理', desc: 'Go + Gin 负责认证、聊天、文档、OCR 任务与知识库接口。' },
  { index: '03', title: '数据分层保存', desc: 'MySQL 保存用户、会话、文档正文与 OCR 任务，Redis 保存 RAG 向量索引。' },
  { index: '04', title: '异步 OCR 消费', desc: '上传后任务进入 RabbitMQ，worker 消费并调用外部 PaddleOCR API。' }
]

const stack = ['Go', 'Gin', 'MySQL', 'Redis Vector', 'RabbitMQ', 'Docker Compose', 'Nginx HTTPS', 'PaddleOCR API']

export default {
  name: 'MenuView',
  components: { ChatDotRound, Document, Search },
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
    return { text, metrics, flow, stack, handleLogout }
  }
}
</script>

<style scoped>
.home-shell { min-height: 100vh; color: #182230; background: #f6f8fb; }
.top-nav { position: sticky; top: 0; z-index: 10; min-height: 72px; display: flex; align-items: center; justify-content: space-between; gap: 18px; padding: 12px clamp(18px, 4vw, 56px); background: rgba(255, 255, 255, 0.92); border-bottom: 1px solid #e6eaf0; backdrop-filter: blur(14px); }
.brand { display: flex; align-items: center; gap: 12px; min-width: 0; }
.brand-mark { width: 42px; height: 42px; display: inline-flex; align-items: center; justify-content: center; border-radius: 10px; background: #172033; color: #fff; font-weight: 800; font-size: 22px; }
.brand strong, .brand span { display: block; }
.brand strong { font-size: 19px; line-height: 1.2; }
.brand span { color: #667085; font-size: 13px; margin-top: 3px; }
.nav-actions { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; justify-content: flex-end; }
.home-main { width: min(1180px, calc(100% - 36px)); margin: 0 auto; padding: 34px 0 56px; }
.hero { display: grid; grid-template-columns: minmax(0, 1.45fr) minmax(280px, .55fr); gap: 26px; align-items: stretch; padding: clamp(28px, 5vw, 58px); border-radius: 8px; background: linear-gradient(135deg, #ffffff 0%, #edf4f2 52%, #f7efe4 100%); border: 1px solid #e2e8f0; box-shadow: 0 18px 48px rgba(20, 32, 50, 0.08); }
.eyebrow, .section-heading p { margin: 0 0 10px; color: #0f766e; font-size: 13px; font-weight: 700; letter-spacing: 0; }
.hero h1 { max-width: 780px; margin: 0; color: #111827; font-size: clamp(34px, 5vw, 60px); line-height: 1.06; letter-spacing: 0; }
.hero-desc { max-width: 760px; margin: 20px 0 0; color: #475467; font-size: 18px; line-height: 1.8; }
.hero-actions { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 28px; }
.status-panel { display: grid; gap: 12px; align-content: end; }
.metric { padding: 18px; border-radius: 8px; background: rgba(255, 255, 255, 0.76); border: 1px solid #dde5ee; }
.metric span { display: block; color: #667085; font-size: 13px; margin-bottom: 8px; }
.metric strong { color: #172033; font-size: 20px; }
.section { margin-top: 34px; }
.section-heading { margin-bottom: 18px; }
.section-heading.compact { margin-bottom: 16px; }
.section-heading h2, .stack-panel h2 { margin: 0; color: #111827; font-size: 28px; line-height: 1.25; }
.feature-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 18px; }
.feature-card { min-height: 232px; padding: 26px; border-radius: 8px; background: #fff; border: 1px solid #e4e9f0; box-shadow: 0 12px 28px rgba(20, 32, 50, 0.06); cursor: pointer; transition: transform .18s ease, box-shadow .18s ease, border-color .18s ease; }
.feature-card:hover { transform: translateY(-4px); border-color: #8bb8ad; box-shadow: 0 18px 34px rgba(20, 32, 50, 0.1); }
.feature-card .el-icon { width: 44px; height: 44px; margin-bottom: 24px; border-radius: 8px; background: #edf4f2; color: #0f766e; font-size: 24px; }
.feature-card h3, .flow-step h3 { margin: 0 0 10px; color: #172033; font-size: 20px; }
.feature-card p, .flow-step p { margin: 0; color: #5d6b7a; line-height: 1.7; }
.split-section { display: grid; grid-template-columns: minmax(0, 1.35fr) minmax(280px, .65fr); gap: 24px; align-items: start; }
.flow-list { display: grid; gap: 12px; }
.flow-step { display: grid; grid-template-columns: 54px 1fr; gap: 14px; padding: 18px; border-radius: 8px; background: #fff; border: 1px solid #e4e9f0; }
.flow-step > span { width: 42px; height: 42px; display: inline-flex; align-items: center; justify-content: center; border-radius: 8px; background: #172033; color: #fff; font-weight: 700; }
.stack-panel { padding: 24px; border-radius: 8px; background: #172033; color: #fff; }
.stack-panel h2 { color: #fff; }
.stack-list { display: flex; flex-wrap: wrap; gap: 10px; margin-top: 20px; }
.stack-list span { padding: 9px 12px; border-radius: 8px; background: rgba(255, 255, 255, 0.1); border: 1px solid rgba(255, 255, 255, 0.16); color: #edf2f7; font-size: 14px; }

@media (max-width: 900px) {
  .top-nav { position: static; align-items: flex-start; }
  .nav-actions { width: 100%; justify-content: flex-start; }
  .home-main { width: min(100% - 28px, 680px); padding-top: 20px; }
  .hero, .split-section { grid-template-columns: 1fr; }
  .hero { padding: 28px; }
  .hero-desc { font-size: 16px; }
  .status-panel { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .feature-grid { grid-template-columns: 1fr; }
  .feature-card { min-height: auto; }
}

@media (max-width: 520px) {
  .top-nav { padding: 12px; gap: 12px; }
  .brand-mark { width: 38px; height: 38px; }
  .brand span { font-size: 12px; }
  .nav-actions { display: grid; grid-template-columns: 1fr 1fr; }
  .nav-actions .el-button { margin-left: 0; min-height: 38px; }
  .nav-actions .el-button:last-child { grid-column: 1 / -1; }
  .home-main { width: calc(100% - 20px); padding: 12px 0 36px; }
  .hero { padding: 22px; }
  .hero h1 { font-size: 32px; }
  .hero-actions { display: grid; grid-template-columns: 1fr; }
  .hero-actions .el-button { margin-left: 0; }
  .status-panel { grid-template-columns: 1fr; }
  .section-heading h2, .stack-panel h2 { font-size: 24px; }
  .feature-card, .flow-step, .stack-panel { padding: 18px; }
  .flow-step { grid-template-columns: 1fr; }
}
</style>
