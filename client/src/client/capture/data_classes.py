from dataclasses import dataclass
import cv2

@dataclass
class Frame:
    img: cv2.UMat
    timestamp: int
