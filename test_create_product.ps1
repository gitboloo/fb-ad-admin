$body = @{
    name = "Digital collectible card game"
    description = "A digital collectible card game"
    logo = "http://localhost:8089/uploads/products/1732097629234-Pokémon_Trading_Card_Game_Pocket_Icon.webp"
    type = "Digital collectible card game"
    banner = "http://localhost:8089/uploads/products/1732097647779-screenshot1.jpg"
    downloads = 0
    price = 0
    screenshots = @(
        "http://localhost:8089/uploads/products/1732097647779-screenshot1.jpg"
    )
    video_url = ""
    status = 1
} | ConvertTo-Json

$headers = @{
    'Content-Type' = 'application/json'
    'Authorization' = 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbl9pZCI6MiwiZXhwIjoxNzMyMTg0MDY5LCJpYXQiOjE3MzIwOTc2Njl9.YUJY5i6AkPUJDe-rPRwaNKZSXKdvvpPOm7mR7Ql1LKI'
}

try {
    $result = Invoke-RestMethod -Uri 'http://localhost:8089/api/admin/products' -Method POST -Headers $headers -Body $body
    Write-Host "成功创建产品!" -ForegroundColor Green
    $result | ConvertTo-Json -Depth 5
} catch {
    Write-Host "创建失败:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    Write-Host $_.Exception.Response
}
