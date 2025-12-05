Recomendaciones por WebSocket (WS)

Este frontend obtiene recomendaciones mediante WebSocket desde el endpoint `GET /ws/recommend/{userId}`.

Cómo probar localmente:

- El cliente está configurado para usar `localhost` durante las pruebas locales. La URL WebSocket por defecto es:

```
ws://localhost:8080/ws/recommend/{userId}?limit=..&genre=..
```

- Abre DevTools → Network → WS para ver la conexión y los mensajes entrantes. Si el backend envía `{ movies, metrics }`, el cliente mostrará `movies` y guardará `metrics` en almacenamiento local.

Notas:
- El proxy REST `app/api/recommend/[userId]/route.ts` permanece para compatibilidad con integraciones que lo consuman por HTTP; al usarlo desde el servidor aparecerá una advertencia en tiempo de ejecución indicando que está deprecado.
