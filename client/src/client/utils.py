import cv2
import time
from client.capture import Frame

def logerr(msg: str):
    print(f"[Error]: {msg}")

def create_frame(img: cv2.typing.MatLike) -> Frame:
    return Frame(img, int(time.time()*1000))
