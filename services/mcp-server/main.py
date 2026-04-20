from domain.google_calendar import CreateSchedule
import httpx
from mcp.server.fastmcp import FastMCP
from contextvars import ContextVar

mcp = FastMCP("bioma-mcp-server")

urls = {
    "google_calendar": "https://www.googleapis.com/calendar/v3/calendars/primary/events",
}

@mcp.tool()
async def schedule_google_calendar(
    summary: str,
    start_datetime: str,
    end_datetime: str,
    access_token: str,
    description: str = None,
    timezone: str = "America/Sao_Paulo",
) -> str:
    """Cria um evento no Google Calendar do usuário.
 
    Args:
        summary: Título do evento.
        start_datetime: Data/hora de início no formato ISO 8601 (ex: 2026-04-21T10:00:00).
        end_datetime: Data/hora de término no formato ISO 8601 (ex: 2026-04-21T11:00:00).
        access_token: Token de acesso OAuth2 do usuário (injetado automaticamente).
        description: Descrição opcional do evento.
        timezone: Fuso horário do evento (padrão: America/Sao_Paulo).
    """
    headers = {
        "Authorization": f"Bearer {access_token}",
        "Content-Type": "application/json",
    }
    body = {
        "summary": summary,
        "start": {"dateTime": start_datetime, "timeZone": timezone},
        "end": {"dateTime": end_datetime, "timeZone": timezone},
    }
    if description:
        body["description"] = description
 
    async with httpx.AsyncClient() as http:
        try:
            resp = await http.post(urls["google_calendar"], headers=headers, json=body)
            resp.raise_for_status()
            link = resp.json().get("htmlLink", "link indisponível")
            return f"Evento criado com sucesso! Acesse aqui: {link}"
        except httpx.HTTPStatusError as e:
            msg = e.response.json().get("error", {}).get("message", str(e))
            return f"Erro ao criar evento no Google Calendar: {msg}"
        except Exception as e:
            return f"Erro interno ao comunicar com o Google Calendar: {e}"
 
 
 
@mcp.tool()
async def schedule_recurring_google_calendar(
    summary: str,
    start_time: str,
    end_time: str,
    weekday: list[str],
    start_date: str,
    end_date: str,
    access_token: str,
    description: str = None,
    timezone: str = "America/Sao_Paulo",
) -> str:
    """Cria um evento recorrente semanal no Google Calendar usando RRULE — um único request para o ano inteiro.

    Args:
        summary: Título do evento.
        start_time: Hora de início no formato HH:MM (ex: 10:00).
        end_time: Hora de término no formato HH:MM (ex: 11:00).
        weekday: Dia da semana em inglês (monday, tuesday, ..., sunday).
        start_date: Data da primeira ocorrência no formato YYYY-MM-DD.
        end_date: Data da última ocorrência no formato YYYY-MM-DD.
        access_token: Token de acesso OAuth2 do usuário (injetado automaticamente).
        description: Descrição opcional.
        timezone: Fuso horário (padrão: America/Sao_Paulo).
    """
    # RRULE usa abreviações de 2 letras em maiúsculo
    WEEKDAY_MAP = {
        "monday": "MO", "tuesday": "TU", "wednesday": "WE",
        "thursday": "TH", "friday": "FR", "saturday": "SA", "sunday": "SU",
    }

    byday = WEEKDAY_MAP.get(weekday.lower())
    if not byday:
        return f"Dia inválido: '{weekday}'. Use: monday, tuesday, wednesday, thursday, friday, saturday, sunday."

    # UNTIL precisa estar no formato YYYYMMDD
    until = end_date.replace("-", "")

    headers = {
        "Authorization": f"Bearer {access_token}",
        "Content-Type": "application/json",
    }

    body = {
        "summary": summary,
        "start": {"dateTime": f"{start_date}T{start_time}:00", "timeZone": timezone},
        "end":   {"dateTime": f"{start_date}T{end_time}:00",   "timeZone": timezone},
        "recurrence": [
            f"RRULE:FREQ=WEEKLY;BYDAY={byday};UNTIL={until}"
        ],
    }
    if description:
        body["description"] = description

    async with httpx.AsyncClient() as http:
        try:
            resp = await http.post(urls["google_calendar"], headers=headers, json=body)
            resp.raise_for_status()
            link = resp.json().get("htmlLink", "link indisponível")
            return f"Evento recorrente criado com sucesso! Toda {weekday} de {start_date} até {end_date}. Link: {link}"
        except httpx.HTTPStatusError as e:
            msg = e.response.json().get("error", {}).get("message", str(e))
            return f"Erro ao criar evento recorrente: {msg}"
        except Exception as e:
            return f"Erro interno: {e}"
 
 
@mcp.tool()
async def list_google_calendar_events(
    access_token: str,
    max_results: int = 10,
) -> str:
    """Lista os próximos eventos do Google Calendar do usuário.
 
    Args:
        access_token: Token de acesso OAuth2 do usuário (injetado automaticamente).
        max_results: Quantidade máxima de eventos a retornar (padrão: 10).
    """
    from datetime import datetime, timezone
 
    headers = {"Authorization": f"Bearer {access_token}"}
    params = {
        "timeMin": datetime.now(timezone.utc).isoformat(),
        "maxResults": max_results,
        "singleEvents": True,
        "orderBy": "startTime",
    }
 
    async with httpx.AsyncClient() as http:
        try:
            resp = await http.get(urls["google_calendar"], headers=headers, params=params)
            resp.raise_for_status()
            events = resp.json().get("items", [])
            if not events:
                return "Nenhum evento encontrado nos próximos dias."
 
            lines = []
            for e in events:
                start = e.get("start", {}).get("dateTime") or e.get("start", {}).get("date", "")
                lines.append(f"- {e.get('summary', 'Sem título')} | {start}")
            return "\n".join(lines)
        except httpx.HTTPStatusError as e:
            msg = e.response.json().get("error", {}).get("message", str(e))
            return f"Erro ao listar eventos: {msg}"
        except Exception as e:
            return f"Erro interno: {e}"
   

@mcp.tool()
def save_google_drive():
    """Funcao responsavel por salvar o arquivo no google drive"""
    return "Funcionalidade de salvar no Google Drive ainda não implementada."


if __name__ == "__main__":
    mcp.run(transport="http", port=8000)