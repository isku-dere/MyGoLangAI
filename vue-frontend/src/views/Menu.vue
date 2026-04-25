<template>
  <div class="menu-container">
    <el-header class="header">
      <h1>AI????</h1>
      <el-button type="danger" @click="handleLogout">????</el-button>
    </el-header>
    <el-main class="main">
      <div class="menu-grid">
        <el-card class="menu-item" @click="$router.push('/ai-chat')">
          <div class="card-content">
            <el-icon size="48" color="#409eff"><ChatDotRound /></el-icon>
            <h3>AI??</h3>
            <p>?AI??????</p>
          </div>
        </el-card>
        <el-card class="menu-item" @click="$router.push('/image-recognition')">
          <div class="card-content">
            <el-icon size="48" color="#67c23a"><Camera /></el-icon>
            <h3>????</h3>
            <p>??????AI??</p>
          </div>
        </el-card>
        <el-card class="menu-item" @click="$router.push('/ocr-notes')">
          <div class="card-content">
            <el-icon size="48" color="#d98146"><Document /></el-icon>
            <h3>OCR ??</h3>
            <p>?????????????? Markdown</p>
          </div>
        </el-card>
      </div>
    </el-main>
  </div>
</template>

<script>
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ChatDotRound, Camera, Document } from '@element-plus/icons-vue'

export default {
  name: 'MenuView',
  components: {
    ChatDotRound,
    Camera,
    Document
  },
  setup() {
    const router = useRouter()

    const handleLogout = async () => {
      try {
        await ElMessageBox.confirm('?????????', '??', {
          confirmButtonText: '??',
          cancelButtonText: '??',
          type: 'warning'
        })
        localStorage.removeItem('token')
        ElMessage.success('??????')
        router.push('/login')
      } catch {
        // ??????
      }
    }

    return {
      handleLogout
    }
  }
}
</script>

<style scoped>
.menu-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;
}

.menu-container::before {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 18% 25%, rgba(255,255,255,.16), transparent 24%), radial-gradient(circle at 82% 70%, rgba(255,255,255,.12), transparent 26%);
  animation: grainMove 30s linear infinite;
}

@keyframes grainMove {
  0% { transform: translate(0, 0); }
  100% { transform: translate(80px, 80px); }
}

.header {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 30px;
  box-shadow: 0 2px 20px rgba(0, 0, 0, 0.1);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  position: relative;
  z-index: 2;
}

.header h1 {
  margin: 0;
  font-size: 28px;
  font-weight: 600;
  color: white;
}

.el-button {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  transition: all 0.3s ease;
}

.el-button:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.2);
}

.main {
  flex: 1;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
  z-index: 1;
}

.menu-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 32px;
  max-width: 1120px;
  width: 100%;
  padding: 40px;
  animation: gridFadeIn 1s ease-out;
}

@keyframes gridFadeIn {
  from { opacity: 0; transform: translateY(50px); }
  to { opacity: 1; transform: translateY(0); }
}

.menu-item {
  cursor: pointer;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(15px);
  border-radius: 20px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  position: relative;
  overflow: hidden;
  animation: cardSlideIn 0.8s ease-out both;
}

.menu-item:nth-child(1) { animation-delay: 0.1s; }
.menu-item:nth-child(2) { animation-delay: 0.2s; }
.menu-item:nth-child(3) { animation-delay: 0.3s; }

@keyframes cardSlideIn {
  from { opacity: 0; transform: translateY(60px) rotateX(10deg); }
  to { opacity: 1; transform: translateY(0) rotateX(0deg); }
}

.menu-item:hover {
  transform: translateY(-15px) scale(1.04);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.card-content {
  text-align: center;
  padding: 46px 26px;
  position: relative;
  z-index: 1;
}

.el-icon {
  display: block;
  margin: 0 auto 20px;
  transition: all 0.3s ease;
}

.menu-item:hover .el-icon {
  transform: scale(1.2) rotate(5deg);
}

.card-content h3 {
  margin: 0 0 15px 0;
  color: #2c3e50;
  font-size: 24px;
  font-weight: 600;
}

.card-content p {
  margin: 0;
  color: #7f8c8d;
  font-size: 16px;
  line-height: 1.6;
}
</style>
