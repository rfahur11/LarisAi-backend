# Azure Container Apps One-Click Deployment Script for LarisAI POS Service
# Sesuai aturan: FinOps Guardrail $100 Student Cap Protection (Scale-to-zero, region southeastasia)

param(
    [string]$ResourceGroup = "rg-larisai-prod",
    [string]$Location = "southeastasia",
    [string]$MongoUri = "mongodb+srv://rfahrur6045_db_user:admin987@cluster0.2q0huva.mongodb.net/?appName=Cluster0",
    [string]$ContainerImage = ""
)

Write-Host "🚀 Memulai Deployment LarisAI ke Azure Container Apps..." -ForegroundColor Cyan

# 1. Cek Login Azure
$account = az account show --output json | ConvertFrom-Json
if (-not $account) {
    Write-Error "❌ Belum login ke Azure CLI. Silakan jalankan 'az login'."
    exit 1
}
Write-Host "✅ Menggunakan Subscription: $($account.name) ($($account.id))" -ForegroundColor Green

# 2. Buat Resource Group jika belum ada
Write-Host "📦 Memastikan Resource Group '$ResourceGroup' di '$Location'..." -ForegroundColor Yellow
az group create --name $ResourceGroup --location $Location --output none

# 3. Validasi Template Bicep
Write-Host "🔍 Menjalankan pre-flight validation (FinOps safe)..." -ForegroundColor Yellow
$bicepPath = Join-Path $PSScriptRoot "azure-container-apps.bicep"

az deployment group validate `
    --resource-group $ResourceGroup `
    --template-file $bicepPath `
    --parameters location=$Location mongoUri=$MongoUri

if ($LASTEXITCODE -ne 0) {
    Write-Error "❌ Validasi template Bicep gagal."
    exit 1
}

Write-Host "✅ Validasi Bicep sukses! Siap deploy dengan scale-to-zero serverless." -ForegroundColor Green
