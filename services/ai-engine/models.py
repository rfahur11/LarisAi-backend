from pydantic import BaseModel, Field
from typing import List

class StockoutPrediction(BaseModel):
    product_id: str
    name: str
    current_stock: int
    daily_burn_rate: float
    days_until_stockout: int

class StockoutResponse(BaseModel):
    predictions: List[StockoutPrediction]

class CustomerCluster(BaseModel):
    customer_id: str
    cluster_label: str
    frequency_count: int
    monetary_value: float

class ClusterResponse(BaseModel):
    silhouette_score: float = 0.0
    clusters: List[CustomerCluster]

class PromoRequest(BaseModel):
    cluster_label: str
    message: str
    channel: str = "whatsapp"

class PromoResponse(BaseModel):
    success: bool
    messages_sent: int
