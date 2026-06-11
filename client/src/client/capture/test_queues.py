from typing import Callable
import time
import cv2
from threading import Thread
from client.capture import Frame, FBuf
from client.utils import logerr
from client.capture.detect_rails import preprocess

raw_buffer = FBuf(val_type=Frame, maxsize=5)
gray_buffer = FBuf(val_type=Frame, maxsize=5)

def transport(buf1: FBuf, processor: Callable, buf2: FBuf, error_callback: Callable = lambda t: logerr(f"Obj from first buffer with type {t.__name__} is None")):
    while True:
        obj1 = buf1.get()
        if obj1 is None:
            error_callback(buf1.val_type)
        else:
            obj2 = processor(obj1)
            if not isinstance(obj2, buf2.val_type):
                raise TypeError(f"Type of object2 is not {buf2.val_type}")
            else:
                buf2.put(obj2)

def to_gray(img: Frame) -> Frame:
    return Frame(cv2.cvtColor(img.img, cv2.COLOR_RGB2GRAY), img.timestamp)

def reading_frames():
    vc = cv2.VideoCapture("data/train/rails1.mp4")
    while True:
        ret, frame = vc.read()
        if ret:
            raw_buffer.put(Frame(frame, int(time.time()*1000)))
    vc.release()

def show_frames():
    cv2.namedWindow("Camera", cv2.WINDOW_NORMAL)
    cv2.resizeWindow("Camera", 800, 600)
    while True:
        frame_obj: Frame = gray_buffer.get()
        if frame_obj is not None:
            cv2.imshow("Camera", frame_obj.img)
        if cv2.waitKey(1) == ord('q'):
            break
    cv2.destroyAllWindows()

read_thread = Thread(target=reading_frames)
gray_thread = Thread(target=transport, args=(raw_buffer, preprocess, gray_buffer))
show_thread = Thread(target=show_frames)

gray_thread.start()
read_thread.start()
show_thread.start()

gray_thread.join()
read_thread.join()
show_thread.join()
