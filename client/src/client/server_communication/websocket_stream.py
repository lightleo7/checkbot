import asyncio
import cv2
import websockets
import numpy
from capture.frame_buffer import FBuf

ws_url = "ws://localhost:8080/ws/stream"

frame_queue = FBuf(numpy.ndarray, 2)


async def ws_streaming_task():
    while True:
        try:
            print("[WS Client] Подключение к WebSocket серверу...")
            async with websockets.connect(ws_url) as websocket:
                print("[WS Client] Успешно подключено к стриму!")

                while True:
                    frame = await asyncio.to_thread(frame_queue.get)
                    
                    if frame is None:
                        await asyncio.sleep(0.01)
                        continue

                    success, encoded_img = cv2.imencode(
                        ".jpg", frame, [int(cv2.IMWRITE_JPEG_QUALITY), 80]
                    )

                    if success:
                        await websocket.send(encoded_img.tobytes())

        except (
            websockets.exceptions.ConnectionClosed,
            ConnectionRefusedError,
            OSError,
        ):
            print(
                "[WS Client] Сервер недоступен. Повторное подключение через 3 секунды..."
            )
            await asyncio.sleep(3)
        except Exception as e:
            print(f"[WS Client] Ошибка стримера: {e}")
            await asyncio.sleep(3)
