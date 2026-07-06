package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

// InitMongoDB menginisialisasi koneksi pooling MongoDB secara stabil dan mutakhir.
func InitMongoDB() *mongo.Database {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "larisai_pos"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Konfigurasi connection pooling
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetMaxPoolSize(50)
	clientOptions.SetMinPoolSize(10)
	clientOptions.SetMaxConnIdleTime(5 * time.Minute)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("❌ Gagal menghubungkan ke MongoDB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("⚠️ Peringatan: MongoDB ping gagal (pastikan MongoDB berjalan di %s): %v", uri, err)
	} else {
		log.Println("✅ Berhasil terhubung ke MongoDB pooling!")
	}

	DB = client.Database(dbName)
	return DB
}
