import asyncio, httpx, cv2
from datetime import datetime

base_url = "http://localhost:8080"

async def UploadToServer(data, files):
    async with httpx.AsyncClient() as client:
        try:
            response = await client.post(f"{base_url}/api/defects", data=data, files=files, timeout=5.0)
            response.raise_for_status()
            return True
            
        except httpx.HTTPError as e:
            print(f"Ошибка при отправке: {e}")
            return False 

async def SendData(Type, Coordinates, cvImages):
    delay = 1

    ready_to_send = {}
    
    i = 1
    for img in cvImages:
        success, encoded_img = cv2.imencode('.jpg', img)
        
        if success:
            img_bytes = encoded_img.tobytes()
            
            ready_to_send[str(i)] = (f"{i}.jpg", img_bytes, "image/jpeg")
        else:
            print(f"Не удалось сжать изображение: {i}")

        i+=1

    data = {
        "Type": Type,
        "Coordinates": str(Coordinates),
        "TimeSpotted": datetime.now().strftime("%H:%M:%S")
    }

    while True:
        if not await UploadToServer(data, ready_to_send):
            print("goyda ne otpravlena")
            await asyncio.sleep(delay)
        else:
            print("goyda otpravlena")
            break

