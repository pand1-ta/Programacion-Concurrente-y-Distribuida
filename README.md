# 🎬 Sistema de Recomendación Distribuido — Filtrado Colaborativo en Go

**Universidad Peruana de Ciencias Aplicadas (UPC)**

**Materia:** Programación Concurrente y Distribuida (CC65)

**Profesor:** Carlos Alberto Jara García  

## 👨‍💻 Integrantes

- Jimena Quintana Noa - `U20201F576`  
- Eduardo Rivas - `UAD266EW`

---

# 🧠 Descripción General del Proyecto

Este proyecto implementa un **sistema de recomendación de películas distribuido**, basado en *filtrado colaborativo* y desarrollado en *Go (Golang)*.
El sistema procesa reseñas de usuarios de manera *concurrente*, utilizando *goroutines y channels* para lograr **menores tiempos de respuesta y alta escalabilidad**.

La propuesta busca demostrar cómo los principios de *programación concurrente y distribuida* pueden mejorar el rendimiento en tareas intensivas de cómputo, como el cálculo de similitudes entre usuarios para la generación de recomendaciones personalizadas.

---

## 🎯 Objetivo del proyecto

Desarrollar e implementar un sistema de recomendación distribuido basado en filtrado colaborativo, capaz de procesar reseñas en paralelo y ofrecer recomendaciones personalizadas con bajo tiempo de respuesta y alta escalabilidad.

---

## 🧩 Arquitectura general

El sistema se compone de tres fases principales:

| Etapa                                           | Descripción                                                                                                                                       |
| :---------------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------ |
| **1️⃣ Preprocesamiento de datos**               | Limpieza y validación de registros del dataset MovieLens. Se eliminan duplicados, se corrigen tipos de datos y se obtienen métricas del conjunto. |
| **2️⃣ Generación de matriz usuario–película**   | Conversión de las calificaciones limpias en una matriz normalizada. Cada fila representa un usuario y cada columna una película.                  |
| **3️⃣ Cálculo concurrente de similitud coseno** | División del cálculo en submatrices procesadas en paralelo mediante goroutines y channels, aplicando el patrón **Fan-Out / Fan-In**.              |

---

## ⚙️ Tecnologías utilizadas

- Lenguaje: Go (Golang)
- Dataset: MovieLens 20M 
- Paradigma: Programación concurrente y distribuida
- Métricas: Tiempo secuencial, tiempo paralelo, speedup y escalabilidad

---

## 🧠 Algoritmo principal

El modelo implementa Filtrado Colaborativo Basado en Usuarios (User-based CF), empleando la Similitud Coseno como métrica para identificar usuarios con gustos similares.

**Fórmula:**

<img width="543" height="147" alt="image" src="https://github.com/user-attachments/assets/f9ba5fe4-ff43-46fd-9416-cac7c2a9cb09" />

---

## 🚀 Resultados de rendimiento

| Tamaño del dataset | Goroutines | Tiempo secuencial | Tiempo paralelo | Speedup |
| :----------------: | :--------: | :---------------: | :-------------: | :-----: |
|       10 000       |      2     |      11.51 ms     |     17.67 ms    |  0.65×  |
|       10 000       |      4     |      11.51 ms     |     11.90 ms    |  0.97×  |
|       10 000       |      8     |      11.51 ms     |     8.03 ms     |  1.43×  |
|       50 000       |      2     |     358.30 ms     |    406.14 ms    |  0.88×  |
|       50 000       |      4     |     358.30 ms     |    282.24 ms    |  1.27×  |
|       50 000       |      8     |     358.30 ms     |    192.69 ms    |  1.86×  |
|       100 000      |      2     |       1.81 s      |      1.80 s     |  1.00×  |
|       100 000      |      4     |       1.81 s      |      1.17 s     |  1.54×  |
|       100 000      |      8     |       1.81 s      |      0.77 s     |  2.35×  |
|       250 000      |      2     |      26.93 s      |     31.78 s     |  0.85×  |
|       250 000      |      4     |      26.93 s      |     20.84 s     |  1.29×  |
|       250 000      |      8     |      26.93 s      |     13.34 s     |  2.02×  |
|       500 000      |      2     |      81.27 s      |     90.39 s     |  0.90×  |
|       500 000      |      4     |      81.27 s      |     58.19 s     |  1.40×  |
|       500 000      |      8     |      81.27 s      |     40.80 s     |  1.99×  |

---

## 📈 Análisis de resultados

El gráfico comparativo de tiempos de ejecución evidencia que:

- Para datasets pequeños (<50 000 registros), la **sobrecarga de concurrencia** disminuye el rendimiento.
- A medida que el volumen de datos crece, el procesamiento concurrente **reduce drásticamente los tiempos de ejecución**.
- Con 8 goroutines, el sistema logra un **speedup cercano a 2×** en datasets grandes (≥100 000 registros).
- Esto confirma la **escalabilidad y eficiencia del enfoque distribuido**, cumpliendo con el objetivo del proyecto.

---

## 🏗️ Estructura del repositorio

```go
📁 Programacion-Concurrente-y-Distribuida
└── PC3/Data/
    ├── ratings.csv
    ├── preprocesamiento.go
    ├── paralelizacion.go
    ├── matriz_usuarios_peliculas.csv
    ├── usuarios_mapping.csv
    └── peliculas_mapping.csv
```
---

## 🧭 Cómo ejecutar el proyecto

1. Clonar el repositorio:
  ```bash
  git clone https://github.com/pand1-ta/Programacion-Concurrente-y-Distribuida.git
  cd Programacion-Concurrente-y-Distribuida/Data
  ```
2. Ejecutar el preprocesamiento:
  ```bash
  go run preprocesamiento.go
  ```
3. Ejecutar la paralelización:
  ```bash
  go run paralelizacion.go
  ```
4. Observar los resultados de tiempos y speedup en consola.

