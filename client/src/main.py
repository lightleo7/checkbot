import asyncio
import cv2
from server_communication.upload import SendData

async def example():
    image = cv2.imread('photo.jpg', cv2.IMREAD_GRAYSCALE)
    image2 = cv2.imread('photo.jpg')
    # one image
    await SendData(Type="treeeeee", Coordinates="12350", cvImages=[image])
    # two images
    await SendData(Type="trees", Coordinates="762381588", cvImages=[image, image2])


async def main():
    cap = cv2.VideoCapture(0)
    
    if not cap.isOpened():
        print("Ошибка: Не удалось открыть камеру!")
        return
    
    frame_count = 0

    try:
        while True:
            ret, frame = cap.read()
            if not ret:
                await asyncio.sleep(0.01)
                continue

            frame_count += 1

            key = cv2.waitKey(1) & 0xFF
            if key == ord('s'):
                asyncio.create_task(SendData(Type="treeeeee", Coordinates="12350", cvImages=[frame]))

            if key == ord('q'):
                print("end...")
                break

            cv2.imshow("framerr", frame)

            await asyncio.sleep(0.001)

    finally:
        cap.release()
        cv2.destroyAllWindows()
        await asyncio.sleep(1)
    

if __name__ == "__main__":
    asyncio.run(main())