import asyncio
import httpx

base_url = "http://172.22.22.186:8080"

async def UploadToServer(data):
    async with httpx.AsyncClient() as client:
        try:
            response = await client.post(f"{base_url}/api/defects", json=data, timeout=5.0)
            
            response.raise_for_status()
            return True
            
        except httpx.HTTPError as e:
            print(f"Ошибка при отправке: {e}")
            return False 

async def SendData(Type, Coordinates, cvImages):
    delay = 1

    data = {
        "Type": Type,
        "Coordinates": int(Coordinates)
    }

    while True:
        if not await UploadToServer(data):
            print("goyda ne otpravlena")
            await asyncio.sleep(delay)
        else:
            break


    ready_to_send = {}
    
    i = 1
    for img in cvImages:
        success, encoded_img = cv2.imencode('.jpg', img)
        
        if success:
            img_bytes = encoded_img.tobytes()
            
            ready_to_send[i] = (f"{i}.jpg", img_bytes, "image/jpeg")
        else:
            print(f"Не удалось сжать изображение: {i}")

        i+=1

    print("goyda otpravlena")



if __name__ == "__main__":
    aw

