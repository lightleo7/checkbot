import asyncio
import cv2
from server_communication.upload import SendData

async def main():
    image = cv2.imread('photo.jpg', cv2.IMREAD_GRAYSCALE)
    await SendData(Type="tree", Coordinates="150", cvImages=[image])

    

if __name__ == "__main__":
    asyncio.run(main())
