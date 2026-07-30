# ====================================================================
# 一键部署脚本 - 本地构建 + 传到虚拟机
#
# 用法：
#   .\deploy.ps1                    # 部署全部（前端+后端）
#   .\deploy.ps1 -Service blades    # 只部署 blades 服务
#   .\deploy.ps1 -Service frontend  # 只部署前端
#   .\deploy.ps1 -Service infra     # 只启动/重启基础设施
#
# 首次使用前，修改下面的 VM_IP、VM_USER、VM_PATH 配置
# ====================================================================

param(
    [string]$Service = "all"
)

# ======================== 虚拟机配置（按需修改）========================
$VM_IP   = "192.168.191.128"
$VM_USER = "cyi3muz"
$VM_PATH = "/home/cyi3muz/cyi-app"
# ====================================================================

$PROJECT_ROOT = $PSScriptRoot
$DEPLOY_DIR   = Join-Path $PROJECT_ROOT "deploy\docker-compose"

# 颜色输出
function Write-Step($msg) { Write-Host "`n========== $msg ==========" -ForegroundColor Cyan }
function Write-OK($msg)   { Write-Host "✅ $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "⚠️ $msg" -ForegroundColor Yellow }
function Write-Err($msg)  { Write-Host "❌ $msg" -ForegroundColor Red }

# SSH 命令快捷方式
function SSH-Run($cmd) {
    ssh "$VM_USER@$VM_IP" $cmd
}

# SCP 传文件到虚拟机
function SCP-Send($local, $remote) {
    scp -r $local "${VM_USER}@${VM_IP}:$remote"
}

# ======================== 服务定义 ========================
# 服务名 → (构建路径, 镜像名, compose服务名)
$Services = @{
    "blades"    = @{ Path = "app\blades";            Image = "cyi3muz/blades:0.1.0";         Compose = "service-blades" }
    "admin"     = @{ Path = "app\cin\admin";         Image = "cyi3muz/cin-admin:0.1.0";      Compose = "service-admin" }
    "interface" = @{ Path = "app\cin\interface";     Image = "cyi3muz/cin-interface:0.1.0";  Compose = "service-interface" }
    "user"      = @{ Path = "app\service\user";      Image = "cyi3muz/service-user:0.1.0";   Compose = "service-user" }
    "job"       = @{ Path = "app\job";               Image = "cyi3muz/job:0.1.0";            Compose = "service-job" }
    "task"      = @{ Path = "app\task";              Image = "cyi3muz/task:0.1.0";           Compose = "service-task" }
}

# ======================== 0. 前置检查 ========================
Write-Step "前置检查"

# 检查 Docker
$dockerOk = docker version 2>$null
if (-not $dockerOk) {
    Write-Err "Docker 未运行，请先启动 Docker Desktop"
    exit 1
}
Write-OK "Docker 正常"

# 测试虚拟机连接
Write-Host "测试虚拟机连接..."
$sshOk = SSH-Run "echo ok" 2>$null
if ($sshOk -ne "ok") {
    Write-Err "无法连接虚拟机 $VM_USER@$VM_IP，请确认 SSH 是否可用"
    Write-Host "提示：如果需要密码登录，请先配置 SSH 密钥免密登录"
    Write-Host "      或者手动执行：ssh $VM_USER@$VM_IP 测试连接"
    exit 1
}
Write-OK "虚拟机连接正常"

# ======================== 1. 同步部署配置文件 ========================
Write-Step "同步部署配置到虚拟机"

# 确保虚拟机项目目录存在
SSH-Run "mkdir -p $VM_PATH/deploy/docker-compose $VM_PATH/deploy/nginx $VM_PATH/web/admin $VM_PATH/web/cyi-mall"

# 传 docker-compose 配置和 .env
SCP-Send "$DEPLOY_DIR\docker-compose.infra.yml" "$VM_PATH/deploy/docker-compose/"
SCP-Send "$DEPLOY_DIR\docker-compose.app.yml"   "$VM_PATH/deploy/docker-compose/"
SCP-Send "$DEPLOY_DIR\.env"                     "$VM_PATH/deploy/docker-compose/"
SCP-Send "$PROJECT_ROOT\deploy\nginx\cyi-app.conf" "$VM_PATH/deploy/nginx/"

Write-OK "配置文件已同步"

# ======================== 2. 基础设施 ========================
if ($Service -eq "all" -or $Service -eq "infra") {
    Write-Step "启动/重启基础设施"
    SSH-Run "cd $VM_PATH/deploy/docker-compose && docker compose -f docker-compose.infra.yml --env-file .env up -d"
    Write-OK "基础设施已启动"

    if ($Service -eq "infra") { exit 0 }
}

# ======================== 3. 前端构建 + 传输 ========================
if ($Service -eq "all" -or $Service -eq "frontend") {
    Write-Step "构建前端"

    # 构建管理后台
    Write-Host "构建管理后台前端..."
    Push-Location "$PROJECT_ROOT\web\admin"
    if (-not (Test-Path "node_modules")) { npm install }
    npm run build
    Pop-Location
    Write-OK "管理后台前端构建完成"

    # 构建用户端商城
    Write-Host "构建用户端商城前端..."
    Push-Location "$PROJECT_ROOT\web\cyi-mall"
    if (-not (Test-Path "node_modules")) { npm install }
    npm run build
    Pop-Location
    Write-OK "用户端商城前端构建完成"

    # 传前端 dist 到虚拟机
    Write-Step "传输前端文件到虚拟机"
    SCP-Send "$PROJECT_ROOT\web\admin\dist" "$VM_PATH/web/admin/"
    SCP-Send "$PROJECT_ROOT\web\cyi-mall\dist" "$VM_PATH/web/cyi-mall/"
    Write-OK "前端文件已传输"

    if ($Service -eq "frontend") {
        # 重启 nginx 容器加载新前端
        SSH-Run "cd $VM_PATH/deploy/docker-compose && docker compose -f docker-compose.app.yml --env-file .env restart nginx 2>/dev/null || true"
        Write-OK "Nginx 已重新加载"
        exit 0
    }
}

# ======================== 4. 微服务构建 + 传输 ========================
$toDeploy = @()
if ($Service -eq "all") {
    $toDeploy = @("blades", "admin", "interface", "user", "job", "task")
} elseif ($Services.ContainsKey($Service)) {
    $toDeploy = @($Service)
} else {
    Write-Err "未知的服务名: $Service"
    Write-Host "可选: all, infra, frontend, blades, admin, interface, user, job, task"
    exit 1
}

foreach ($svc in $toDeploy) {
    $info = $Services[$svc]
    $buildPath = Join-Path $PROJECT_ROOT $info.Path
    $image = $info.Image
    $compose = $info.Compose

    Write-Step "构建 $svc 服务"

    # 本地构建 Docker 镜像
    Write-Host "构建镜像: $image"
    Push-Location $buildPath
    make docker
    if ($LASTEXITCODE -ne 0) {
        Write-Err "构建 $svc 镜像失败"
        Pop-Location
        continue
    }
    Pop-Location
    Write-OK "镜像构建完成: $image"

    # 保存镜像为 tar 并传到虚拟机
    Write-Host "保存镜像并传输到虚拟机..."
    $tarName = "$svc.tar"
    docker save $image -o $tarName
    scp $tarName "${VM_USER}@${VM_IP}:/tmp/$tarName"
    Remove-Item $tarName -Force

    # 虚拟机加载镜像
    Write-Host "虚拟机加载镜像..."
    SSH-Run "docker load -i /tmp/$tarName && rm /tmp/$tarName"
    Write-OK "镜像已加载到虚拟机"
}

# ======================== 5. 启动/重启应用服务 ========================
Write-Step "启动应用服务"

if ($Service -eq "all") {
    # 启动所有应用服务 + Nginx
    SSH-Run "cd $VM_PATH/deploy/docker-compose && docker compose -f docker-compose.app.yml --env-file .env up -d"
} else {
    # 只重启变更的服务
    foreach ($svc in $toDeploy) {
        $compose = $Services[$svc].Compose
        Write-Host "重启 $compose ..."
        SSH-Run "cd $VM_PATH/deploy/docker-compose && docker compose -f docker-compose.app.yml --env-file .env up -d --no-deps $compose"
    }
    # 重启 nginx（前端可能更新了）
    SSH-Run "cd $VM_PATH/deploy/docker-compose && docker compose -f docker-compose.app.yml --env-file .env restart nginx 2>/dev/null || true"
}

Write-OK "应用服务已启动"

# ======================== 6. 查看状态 ========================
Write-Step "服务状态"
SSH-Run "cd $VM_PATH/deploy/docker-compose && echo '--- 基础设施 ---' && docker compose -f docker-compose.infra.yml --env-file .env ps && echo '' && echo '--- 应用服务 ---' && docker compose -f docker-compose.app.yml --env-file .env ps"

Write-Host ""
Write-OK "部署完成！"
Write-Host ""
Write-Host "访问地址（需配置 hosts 文件）：" -ForegroundColor Cyan
Write-Host "  管理后台:  http://admin.cyi.local"
Write-Host "  用户商城:  http://mall.cyi.local"
Write-Host "  Consul:    http://$VM_IP`:8500"
Write-Host "  Grafana:   http://$VM_IP`:3000"
Write-Host ""
Write-Host "hosts 文件路径: C:\Windows\System32\drivers\etc\hosts" -ForegroundColor Yellow
Write-Host "  $VM_IP  admin.cyi.local"
Write-Host "  $VM_IP  mall.cyi.local"
