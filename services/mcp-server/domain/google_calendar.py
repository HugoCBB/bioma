from pydantic import BaseModel
from typing import Optional

class EventDateTime(BaseModel):
    dateTime: Optional[str] = None
    date: Optional[str] = None
    timeZone: Optional[str] = "America/Sao_Paulo"

class CreateSchedule(BaseModel):
    summary: str
    description: Optional[str] = None
    start: EventDateTime
    end: EventDateTime