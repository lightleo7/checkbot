import asyncio
import cv2
from defects import DefectTypes
from server_communication.upload import send_data
from server_communication.websocket_stream import ws_streaming_task, frame_queue


async def example():
    image = cv2.imread("photo.jpg", cv2.IMREAD_GRAYSCALE)
    image2 = cv2.imread("photo.jpg")
    # one image
    await send_data(Type=DefectTypes.TREE, Coordinates="12350", cvImages=[image])
    # two images
    await send_data(
        Type=DefectTypes.PERSON, Coordinates="762381588", cvImages=[image, image2]
    )


async def main():
    cap = cv2.VideoCapture(0)

    if not cap.isOpened():
        print("Ошибка: Не удалось открыть камеру!")
        return

    frame_count = 0
    STREAM_EVERY_N_FRAME = 3

    ws_task = asyncio.create_task(ws_streaming_task())

    try:
        while True:
            ret, frame = cap.read()
            if not ret:
                await asyncio.sleep(0.01)
                continue

            frame_count += 1

            key = cv2.waitKey(1) & 0xFF
            
            if key == ord("s"):
                asyncio.create_task(
                    asyncio.to_thread(
                        send_data,
                        Type=DefectTypes.TREE,
                        Coordinates="12350",
                        cvImages=[frame.copy()],
                    )
                )

            if key == ord("q"):
                print("end...")
                break

            cv2.imshow("framerr", frame)

            if frame_count % STREAM_EVERY_N_FRAME == 0:
                frame_queue.put(frame.copy())

            await asyncio.sleep(0.005)

    finally:
        ws_task.cancel()
        cap.release()
        cv2.destroyAllWindows()
        await asyncio.sleep(0.5)


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\nKeyboardInterrupt")
