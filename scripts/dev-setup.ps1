# ====================================================================
# 本地开发一键环境准备（Windows PowerShell 版）
#
# 功能：
#   1. 检查 Docker，必要时从 .env.example 生成 .env
#   2. 启动基础设施（MySQL / Redis / Consul / NATS / Jaeger / RustFS 等）
#   3. 等待 Consul 就绪
#   4. 把各服务 configs/*.yaml 转成本地地址与端口后写入 Consul KV
#
# 用法：
#   powershell -ExecutionPolicy Bypass -File .\scripts\dev-setup.ps1
#   powershell -ExecutionPolicy Bypass -File .\scripts\dev-setup.ps1 -SkipInfra
#
# 说明：
#   - Consul 由 Docker 提供，本机无需安装
#   - 重复执行是安全的，配置会覆盖写入
# ====================================================================

param(
    [switch]$SkipInfra,
    [string]$ConsulAddr = "127.0.0.1:8500"
)

$ErrorActionPreference = "Stop"

$ProjectRoot = Split-Path -Parent $PSScriptRoot
$ComposeDir  = Join-Path $ProjectRoot "deploy\docker-compose"
$AppDir      = Join-Path $ProjectRoot "app"

function Write-Step($msg) { Write-Host "`n==> $msg" -ForegroundColor Cyan }
function Write-OK($msg)   { Write-Host "  [OK] $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "  [..] $msg" -ForegroundColor Yellow }
function Write-Err($msg)  { Write-Host "  [!!] $msg" -ForegroundColor Red }

# --------------------------------------------------------------------
# 1. 检查 Docker 与 Compose
# --------------------------------------------------------------------
Write-Step "检查 Docker"
docker version --format '{{.Server.Version}}' 2>$null | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Err "Docker 未运行，请先启动 Docker Desktop"
    exit 1
}
$ComposeExe = "docker"
$ComposePrefix = @("compose")
docker compose version 2>$null | Out-Null
if ($LASTEXITCODE -ne 0) {
    docker-compose version 2>$null | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Err "未找到 docker compose / docker-compose"
        exit 1
    }
    $ComposeExe = "docker-compose"
    $ComposePrefix = @()
}
Write-OK "Docker 正常"

# --------------------------------------------------------------------
# 2. 准备 .env（端口配置）
# --------------------------------------------------------------------
$EnvFile = Join-Path $ComposeDir ".env"
$EnvExample = Join-Path $ComposeDir ".env.example"
if (-not (Test-Path $EnvFile)) {
    Copy-Item $EnvExample $EnvFile
    Write-OK "已从 .env.example 生成 .env"
}

$EnvMap = @{}
Get-Content $EnvFile | ForEach-Object {
    if ($_ -match '^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$') {
        $EnvMap[$matches[1]] = $matches[2].Trim()
    }
}
function Get-Port($name, $default) {
    if ($EnvMap.ContainsKey($name) -and $EnvMap[$name]) { return $EnvMap[$name] }
    return $default
}

# --------------------------------------------------------------------
# 3. 启动基础设施
# --------------------------------------------------------------------
if (-not $SkipInfra) {
    Write-Step "启动基础设施（首次拉取镜像可能较慢）"
    Push-Location $ComposeDir
    try {
        $composeArgs = $ComposePrefix + @("-f", "docker-compose.infra.yml", "--env-file", ".env", "-p", "cyi-app", "up", "-d")
        & $ComposeExe @composeArgs
        if ($LASTEXITCODE -ne 0) {
            Write-Err "基础设施启动失败，请检查 Docker 日志"
            exit 1
        }
    } finally {
        Pop-Location
    }
    Write-OK "基础设施已启动"
} else {
    Write-Warn "已跳过基础设施启动"
}

# --------------------------------------------------------------------
# 4. 等待 Consul 就绪
# --------------------------------------------------------------------
Write-Step "等待 Consul 就绪"
$consulReady = $false
for ($i = 0; $i -lt 30; $i++) {
    try {
        Invoke-RestMethod -Uri "http://$ConsulAddr/v1/status/leader" -TimeoutSec 2 | Out-Null
        $consulReady = $true
        break
    } catch {
        Start-Sleep -Seconds 2
    }
}
if (-not $consulReady) {
    Write-Err "Consul 未就绪，请执行 docker compose -f deploy/docker-compose/docker-compose.infra.yml ps 检查"
    exit 1
}
Write-OK "Consul 已就绪：http://$ConsulAddr"

# --------------------------------------------------------------------
# 5. 写入公共配置（launch.json 会读取 /local/common）
# --------------------------------------------------------------------
function Put-ConsulConfig($key, $content) {
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($content)
    Invoke-RestMethod -Method Put -Uri "http://$ConsulAddr/v1/kv/$key" -Body $bytes -ContentType "application/octet-stream" | Out-Null
}

Put-ConsulConfig "local/common/config.yaml" "{}"
Write-OK "已写入公共配置 local/common/config.yaml"

# --------------------------------------------------------------------
# 6. 各服务端口映射（容器内端口 -> 本机端口）
#    必须与 deploy/docker-compose/.env 保持一致
# --------------------------------------------------------------------
$PortMaps = @{
    "cin/interface"     = @{ "8000" = (Get-Port "INTERFACE_HTTP_PORT" "8010"); "9000" = (Get-Port "INTERFACE_GRPC_PORT" "9010") }
    "cin/admin"         = @{ "8001" = (Get-Port "ADMIN_HTTP_PORT" "8080");     "9001" = (Get-Port "ADMIN_GRPC_PORT" "9091") }
    "service/user"      = @{ "8000" = "8100"; "9000" = (Get-Port "USER_GRPC_PORT" "9100") }
    "service/order"     = @{ "8003" = (Get-Port "ORDER_HTTP_PORT" "9104"); "9003" = (Get-Port "ORDER_GRPC_PORT" "9101") }
    "service/product"   = @{ "8000" = "8102"; "9000" = (Get-Port "PRODUCT_GRPC_PORT" "9102") }
    "service/marketing" = @{ "8000" = "8103"; "9000" = (Get-Port "MARKETING_GRPC_PORT" "9103") }
    "service/chat"      = @{ "8005" = "8105"; "9005" = (Get-Port "CHAT_GRPC_PORT" "9105") }
    "blades"            = @{ "8000" = (Get-Port "BLADES_HTTP_PORT" "30800"); "9000" = (Get-Port "BLADES_GRPC_PORT" "30900") }
    "task"              = @{ "8000" = (Get-Port "TASK_PORT" "10100"); "9000" = "10101" }
    "nats"              = @{ "38000" = "38000"; "39000" = "39000" }
}

# 容器内主机名 -> 占位符 -> 本机地址。
# 先替换成占位符，避免后续端口替换误伤 OSS 的 9000 等端口。
$HConsul = "127.0.0.1:" + (Get-Port "CONSUL_PORT" "8500")
$HMysql  = "127.0.0.1:" + (Get-Port "MYSQL_PORT" "3306")
$HRedis  = "127.0.0.1:" + (Get-Port "REDIS_PORT" "6379")
$HNats   = "127.0.0.1:" + (Get-Port "NATS_PORT" "4222")
$HJaeger = "127.0.0.1:" + (Get-Port "JAEGER_OTLP_GRPC_PORT" "4318")
$HRustfs = "127.0.0.1:" + (Get-Port "RUSTFS_API_PORT" "9001")
$HDtm    = "http://127.0.0.1:" + (Get-Port "DTM_PORT" "36789")
$HBlades = "127.0.0.1:" + (Get-Port "BLADES_HTTP_PORT" "30800")

$HostPlaceholders = @(
    @("consul:8500", "__HOST_CONSUL__", $HConsul),
    @("mysql:3306",  "__HOST_MYSQL__",  $HMysql),
    @("redis:6379",  "__HOST_REDIS__",  $HRedis),
    @("nats:4222",   "__HOST_NATS__",   $HNats),
    @("jaeger:4318", "__HOST_JAEGER__", $HJaeger),
    @("rustfs:9000", "__HOST_RUSTFS__", $HRustfs),
    @("http://dtm:36789", "__HOST_DTM__", $HDtm),
    @("blades:8000", "__HOST_BLADES__", $HBlades)
)

# --------------------------------------------------------------------
# 7. 遍历各服务 configs/*.yaml，转换后写入 Consul KV
# --------------------------------------------------------------------
Write-Step "导入服务配置到 Consul KV"

# 只导入服务配置白名单，避免把 ss.yaml 等无关文件写入 Consul
$AllowedFiles = @("config.yaml", "auth.yaml", "register.yaml", "transaction.yaml")
$configFiles = Get-ChildItem -Path $AppDir -Recurse -Filter "*.yaml" |
    Where-Object { $_.DirectoryName -match "\\configs$" -and $AllowedFiles -contains $_.Name }

$count = 0
foreach ($file in $configFiles) {
    $rel = $file.Directory.Parent.FullName.Substring($AppDir.Length).TrimStart("\").Replace("\", "/")
    $content = Get-Content $file.FullName -Raw

    # 1) 主机名先换成占位符
    foreach ($pair in $HostPlaceholders) {
        $content = $content.Replace($pair[0], $pair[1])
    }

    # 2) 按服务替换端口
    if ($PortMaps.ContainsKey($rel)) {
        foreach ($src in $PortMaps[$rel].Keys) {
            $content = $content.Replace(":$src", ":" + $PortMaps[$rel][$src])
        }
    } else {
        Write-Warn "未定义端口映射，按原样导入：$rel/$($file.Name)"
    }

    # 3) 占位符换成真实地址
    foreach ($pair in $HostPlaceholders) {
        $content = $content.Replace($pair[1], $pair[2])
    }

    $key = "local/cyi-app/$rel/$($file.Name)"
    Put-ConsulConfig $key $content
    Write-OK "$key"
    $count++
}

Write-Host ""
Write-Host "配置导入完成，共 $count 个文件。" -ForegroundColor Green
Write-Host ""
Write-Host "下一步：" -ForegroundColor Cyan
Write-Host "  1. 在 VS Code 打开「运行和调试」，选择 user_service 等配置按 F5 启动"
Write-Host "  2. Consul 控制台：http://$ConsulAddr/ui"
Write-Host "  3. 若服务启动报配置缺失，重新执行本脚本即可"
Write-Host ""
Write-Host "提示：生产部署请勿使用本脚本，容器内使用服务名访问基础设施。" -ForegroundColor Yellow
