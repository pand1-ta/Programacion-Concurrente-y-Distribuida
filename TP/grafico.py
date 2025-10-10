import matplotlib.pyplot as plt

data_sizes = [1000000, 2000000, 5000000, 10000000, 15000000, 21721689]
t_secuencial = [3.53, 13.02, 103.90, 393.25, 710.92, 1341.08]
t_concurrente_2 = [2.99, 10.02, 87.06, 250.50, 463.23, 910.94]
t_concurrente_4 = [2.53, 9.12, 77.23, 227.17, 395.18, 746.79]
t_concurrente_8 = [3.00, 10.05, 73.99, 244.61, 445.78, 813.95]

try:
    combined_data = sorted(zip(data_sizes, t_secuencial, t_concurrente_2, t_concurrente_4, t_concurrente_8))
    sorted_sizes, sorted_seq, sorted_c2, sorted_c4, sorted_c8 = zip(*combined_data)
except ValueError:
    print("Error: Asegúrate de que todas las listas de datos tengan la misma cantidad de elementos.")
plt.style.use('seaborn-v0_8-whitegrid')
plt.figure(figsize=(12, 7))
# Dibujar cada línea del gráfico
plt.plot(sorted_sizes, sorted_seq, 'o--', color='black', label='Secuencial')
plt.plot(sorted_sizes, sorted_c2, 'o-', color='blue', label='Concurrente (2 goroutines)')
plt.plot(sorted_sizes, sorted_c4, 'o-', color='green', label='Concurrente (4 goroutines)')
plt.plot(sorted_sizes, sorted_c8, 'o-', color='darkorange', label='Concurrente (8 goroutines)')
# Configurar Títulos y Etiquetas
plt.title('Sistema de Recomendación - Comparación: Secuencial vs Concurrente', fontsize=16)
plt.xlabel('Número de Reseñas Procesadas (10M)', fontsize=12)
plt.ylabel('Tiempo de Ejecución (ms)', fontsize=12)

# Formatear el eje X para que los números grandes sean legibles
plt.ticklabel_format(style='sci', axis='x', scilimits=(0,0))

plt.legend(fontsize=11)
plt.grid(True, which='both', linestyle='--', linewidth=0.5)
# Guardar el gráfico en un archivo y mostrarlo en pantalla
output_filename = 'comparacion_rendimiento_manual.png'
plt.savefig(output_filename, dpi=300, bbox_inches='tight')

print(f"- Gráfico guardado exitosamente como '{output_filename}'")

plt.show()