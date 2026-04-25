# GopherAI 公网部署

## 1. 服务器准备
- 系统: Ubuntu 22.04/24.04
- 开放端口: `22`, `80`, `443`
- 安装: Docker + Docker Compose Plugin

## 2. 上传并配置项目
```bash
git clone <your-repo-url>
cd GopherAI-v2/deploy
cp .env.example .env
```

编辑 `.env`，至少填写:
- `OPENAI_API_KEY`
- `OPENAI_MODEL_NAME`
- `OPENAI_BASE_URL`

## 3. 修改域名
编辑 `deploy/nginx/conf.d/gopherai.conf`，把 `api.example.com` 改成你的域名。

## 4. 首次签发 HTTPS 证书
先确保域名 A 记录已指向服务器公网 IP，然后执行:
```bash
sudo certbot certonly --standalone -d your.domain.com
```

证书会生成在 `/etc/letsencrypt/live/your.domain.com/`。

## 5. 启动服务
```bash
docker compose up -d --build
```

## 6. 查看状态
```bash
docker compose ps
docker compose logs -f app
```

## 7. 自动续期证书
```bash
sudo crontab -e
```
添加:
```cron
0 3 * * * certbot renew --quiet && docker compose -f /path/to/GopherAI-v2/deploy/docker-compose.yml restart nginx
```

## 8. 安全建议
- 不要把 `3306/6379/5672` 暴露到公网。
- `.env` 不要提交到仓库。
- 生产环境请替换默认密码。

