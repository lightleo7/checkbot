import asyncio, httpx, cv2
from datetime import datetime

base_url = "http://localhost:8080"
pswd = "1234"

api_token = None

async def GetToken(client: httpx.AsyncClient) -> bool:
    global api_token
    try:
        response = await client.post(f"{base_url}/api/auth", json={"password": pswd}, timeout=5.0)
        if response.status_code == 200:
            api_token = response.json().get("token")
            print("[Python] Токен авторизации успешно получен и сохранен.")
            return True
        else:
            print(f"[Python] Ошибка авторизации: {response.status_code} - {response.text}")
            return False
    except httpx.HTTPError as e:
        print(f"[Python] Сетевая ошибка при попытке авторизации: {e}")
        return False

async def UploadToServer(data, files):
    global api_token
    async with httpx.AsyncClient() as client:
        if not api_token:
            if not await GetToken(client):
                return False

        headers = {"X-API-Token": api_token}
        try:
            response = await client.post(f"{base_url}/api/defects", data=data, headers=headers, files=files, timeout=5.0)

            if response.status_code == 401:
                print("[Python] Токен устарел. Попытка обновить...")
                if await GetToken(client):
                    headers["X-API-Token"] = api_token
                    response = await client.post(f"{base_url}/api/defects", data=data, files=files, headers=headers, timeout=5.0)
                else:
                    return False

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

