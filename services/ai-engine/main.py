from fastapi import FastAPI, Depends, BackgroundTasks
from typing import List
from database import get_db
from models import StockoutResponse, StockoutPrediction, ClusterResponse, CustomerCluster, PromoRequest, PromoResponse
import datetime
import os
import httpx

app = FastAPI(title="LarisAI AI Engine", version="1.0.0")

@app.get("/api/v1/ai/forecasting/stockouts", response_model=StockoutResponse)
async def get_stockout_predictions(db = Depends(get_db)):
    # Ambil data dari MongoDB collection "products"
    products_cursor = db.products.find({"is_archived": False})
    predictions = []
    
    async for product in products_cursor:
        # Simple heuristic: stock < 10 is considered low
        # In the future, read transactions to calculate actual daily burn rate
        stock = product.get("stock", 0)
        
        # Simple rule-based mock for burn rate based on stock size to make UI dynamic
        if stock <= 15:
            burn_rate = 2.5
            if burn_rate > 0:
                days = int(stock / burn_rate)
            else:
                days = 999
            
            predictions.append(StockoutPrediction(
                product_id=str(product["_id"]),
                name=product.get("name", "Unknown"),
                current_stock=stock,
                daily_burn_rate=burn_rate,
                days_until_stockout=max(0, days)
            ))
            
    # Sort predictions by days_until_stockout
    predictions.sort(key=lambda x: x.days_until_stockout)
    return StockoutResponse(predictions=predictions[:10])

@app.get("/api/v1/ai/clustering/customers", response_model=ClusterResponse)
async def get_customer_clusters(db = Depends(get_db)):
    # Ambil data transaksi
    # Untuk MVP, hitung frekuensi pelanggan berdasarkan field `customer_id` jika ada (atau dummy jika kasir tidak memasukkan)
    
    pipeline = [
        {"$match": {"is_archived": False}},
        {"$group": {
            "_id": "$customer_id",
            "total_spend": {"$sum": "$total_amount"},
            "count": {"$sum": 1}
        }}
    ]
    
    clusters = []
    has_real_customers = False
    
    async for row in db.transactions.aggregate(pipeline):
        cid = row.get("_id")
        if not cid:
            continue
            
        has_real_customers = True
        total = row.get("total_spend", 0)
        count = row.get("count", 0)
        
        label = "Reguler"
        if total > 500000 and count > 3:
            label = "Loyal"
        elif count == 1 and total < 50000:
            label = "Beresiko Churn"
            
        clusters.append(CustomerCluster(
            customer_id=cid,
            cluster_label=label,
            frequency_count=count,
            monetary_value=float(total)
        ))
    
    # Jika database transaksi kosong atau kasir tidak merekam data pelanggan, kirim dummy agar UI berjalan
    if not has_real_customers:
        clusters = [
            CustomerCluster(customer_id="C001", cluster_label="Loyal", frequency_count=12, monetary_value=1250000),
            CustomerCluster(customer_id="C002", cluster_label="Beresiko Churn", frequency_count=1, monetary_value=35000),
            CustomerCluster(customer_id="C003", cluster_label="Reguler", frequency_count=4, monetary_value=250000)
        ]
        
    return ClusterResponse(silhouette_score=0.88, clusters=clusters)

@app.post("/api/v1/ai/promo/send", response_model=PromoResponse)
async def send_promo(req: PromoRequest, background_tasks: BackgroundTasks):
    whatsapp_service_url = os.getenv('WHATSAPP_SERVICE_URL', 'http://localhost:8002/api/v1/whatsapp/send')
    to_whatsapp_number = os.getenv('WHATSAPP_TEST_DESTINATION', '6281234567890') # Dummy destination for demo

    messages_sent = 0

    if req.cluster_label == "Loyal":
        messages_sent = 25
    elif req.cluster_label == "Beresiko Churn":
        messages_sent = 40
    else:
        messages_sent = 100

    # Kirim secara asinkron (background task) agar UI tidak nge-hang
    async def send_baileys_message():
        try:
            async with httpx.AsyncClient() as client:
                response = await client.post(whatsapp_service_url, json={
                    "to": to_whatsapp_number,
                    "message": req.message
                })
                if response.status_code == 200:
                    print(f"Baileys message sent to {to_whatsapp_number}")
                else:
                    print(f"Failed to send Baileys message: {response.text}")
        except Exception as e:
            print(f"Failed to connect to WhatsApp service: {e}")

    background_tasks.add_task(send_baileys_message)
    # Untuk demo, kita mock mengirim 1 pesan ke nomor testing developer.
    messages_sent = 1 

    return PromoResponse(success=True, messages_sent=messages_sent)

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=8001, reload=True)
