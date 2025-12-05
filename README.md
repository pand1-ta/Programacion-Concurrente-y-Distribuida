# 🎬 Sistema Distribuido de Recomendación de Películas – CC65 (Entrega Final)

**Materia:** Programación Concurrente y Distribuida (CC65)  
**Profesor:** Carlos Alberto Jara García  

## 👨‍💻 Integrantes

- **Jimena Alexsandra Quintana Noa** – `U20201F576`  
- **Eduardo Rivas** – `uad266ew`

---

# 🧠 Descripción General del Proyecto

Este proyecto implementa un **sistema distribuido de recomendación de películas** basado en filtrado colaborativo entre usuarios, utilizando **Go** para el motor de recomendación y **Next.js** para la interfaz web.

El objetivo central es construir una arquitectura **concurrente, distribuida y escalable**, capaz de procesar grandes volúmenes de datos y generar recomendaciones personalizadas con eficiencia. El sistema incorpora:

- Goroutines para paralelismo  
- Workers TCP para distribución de trabajo  
- Redis para caching  
- MongoDB para persistencia de métricas  
- API REST para comunicación entre frontend y backend  
- Interfaz web moderna en Next.js  

El dataset utilizado es **MovieLens 20M**, mapeado y procesado para calcular similitudes, predecir ratings y mostrar recomendaciones al usuario final.

---

# 🧩 Arquitectura General del Sistema

El sistema está compuesto por varios módulos:

### **1️⃣ Coordinator (Go)**
- Divide tareas entre workers  
- Consolida resultados  
- Coordina el procesamiento distribuido  

### **2️⃣ Workers TCP (Go)**
- Reciben tareas desde el Coordinator  
- Ejecutan cálculos de similitud y predicción  
- Manejan fallos con recuperación segura  

### **3️⃣ API REST (Next.js)**
Endpoints clave:

- `/api/users`  
- `/api/movies`  
- `/api/recommend/:userId`  

Estas rutas actúan como **proxy** hacia el backend distribuido.

### **4️⃣ Redis**
- Cachea resultados para reducir cómputos repetidos.

### **5️⃣ MongoDB**
- Guarda métricas de rendimiento:
  - tiempo secuencial  
  - concurrente  
  - distribuido  
  - workers utilizados  

### **6️⃣ Frontend (Next.js)**
Incluye:

- Login de usuarios  
- Filtros dinámicos  
- Modo general / recomendado  
- Panel de métricas  
- Grid responsivo de películas  

---

# 🎯 Objetivos del Proyecto Final

- Implementar un motor de recomendación distribuido basado en filtrado colaborativo.  
- Construir una arquitectura que combine concurrencia (goroutines) con distribución (Workers TCP).  
- Diseñar una API unificada para desacoplar backend y frontend.  
- Implementar una interfaz web moderna orientada al usuario final.  
- Integrar Redis y MongoDB para caching y persistencia de métricas.  
- Comparar el rendimiento secuencial, concurrente y distribuido para evidenciar mejoras reales.

---

# 🏗️ Componentes Implementados

## 🟦 Backend: Coordinator + Workers (Go)

Características:

- Procesamiento del dataset  
- Cálculo de similitud coseno  
- Predicción de ratings  
- Dividir y enviar chunks a workers  
- Manejo de errores y reconexiones  
- Benchmarking interno  

Ejemplo de cálculo de similitud coseno:

```go
func cosine(u, v []float64) float64 {
    var dot, nu, nv float64
    for i := range u {
        dot += u[i] * v[i]
        nu += u[i] * u[i]
        nv += v[i] * v[i]
    }
    if nu == 0 || nv == 0 {
        return 0
    }
    return dot / (math.Sqrt(nu) * math.Sqrt(nv))
}
```
---

API REST en Next.js  
Ejemplo:  `/api/recommend/[userId]`  

```ts
const BACKEND = process.env.RECOMMENDER_BACKEND || 'http://localhost:8080'

export async function GET(req: Request, context: any) {
  const params = await context.params
  const target = new URL(`${BACKEND}/recommend/${params.userId}`)
  target.search = new URL(req.url).search

  const res = await fetch(target.toString(), { cache: 'no-store' })
  const text = await res.text()

  try {
    return NextResponse.json(JSON.parse(text))
  } catch {
    return new NextResponse(text)
  }
}

```





---

## 🟩 Frontend en Next.js

Incluye:

- Sistema de login por usuario
- Grid de películas (4–5 columnas)
- Modo general / recomendado
- Filtros dinámicos
- Panel emergente de métricas
- Uso de `localStorage ` para sesión
- Estilos en CSS con UI moderna

---

## 🟥 Redis + MongoDB
Redis

- Cachea recomendaciones por usuario.
- Reduce carga del Coordinator + Workers.

MongoDB

Guarda métricas como:

```json
{
  "userId": 123,
  "totalTimeSequential": 1213,
  "totalTimeConcurrent": 422,
  "totalTimeDistributed": 88,
  "workersUsed": 4,
  "timestamp": "2025-12-04"
}

```
---

# 🚀 Ejecución del Sistema

Docker inicia:

```sh
docker compose up --build

```

- Coordinator
- 3+ Workers
- Redis
- MongoDB
- API Next.js
- Frontend

---

#🧩 Conclusiones

El proyecto demuestra el impacto real de combinar concurrencia, distribución y caching en sistemas de recomendación. La arquitectura implementada permite reducir drásticamente los tiempos de cálculo y escalar horizontalmente mediante workers adicionales.
El frontend y la API REST permiten que el sistema sea utilizable por usuarios finales, mientras que Redis y MongoDB añaden robustez, persistencia y rendimiento general.

---

# 💡 Recomendaciones Futuras

- Incorporar balanceo de carga entre workers.
- Migrar protocolo TCP → gRPC para mayor eficiencia.
- Integrar métricas en tiempo real (Prometheus + Grafana).
- Mejorar modelo de recomendación (ALS, embeddings).
- Despliegue completo en la nube.
- Automatizar reinicio y registro de fallos de workers.



