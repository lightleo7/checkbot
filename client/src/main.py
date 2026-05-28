import asyncio
import cv2
from server_communication.upload import SendData

async def example():
    image = cv2.imread('photo.jpg', cv2.IMREAD_GRAYSCALE)
    image2 = cv2.imread('photo.jpg')
    # one image
    await SendData(Type="tree", Coordinates="150", cvImages=[image])
    # two images
    await SendData(Type="trees", Coordinates="762381588", cvImages=[image, image2])


async def main():
    pass
    

if __name__ == "__main__":
    asyncio.run(main())

    # asyncio.run(example())
