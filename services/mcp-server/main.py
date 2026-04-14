from domain.google_calendar import CreateSchedule
import requests
from mcp.server.fastmcp import FastMCP
from contextvars import ContextVar

mcp = FastMCP("bioma-mcp-server")

urls = {
    "google_calendar": "https://www.googleapis.com/calendar/v3/calendars/primary/events",
}

_access_token: ContextVar[str] = ContextVar("access_token", default="")
@mcp.tool()
async def schedule_google_calendar(
    summary: str,
    start_datetime: str,
    end_datetime: str,
    access_token: str,
    description: str = None,
    timezone: str = "America/Sao_Paulo"
):
    
    """Schedule an event in Google Calendar.
    
    Args:
        summary: Event title
        start_datetime: Start datetime in ISO format (e.g. 2026-04-17T10:00:00)
        end_datetime: End datetime in ISO format (e.g. 2026-04-17T11:00:00)
        description: Optional event description
        timezone: Timezone (default: America/Sao_Paulo)
    """

    headers = {
        "Authorization": f"Bearer {access_token}",
        "Content-Type": "application/json"
    }
    
    payload = {
        "summary": summary,
        "start": {"dateTime": start_datetime, "timeZone": timezone},
        "end": {"dateTime": end_datetime, "timeZone": timezone},
    }
    if description:
        payload["description"] = description

    try:
        response = requests.post(urls["google_calendar"], headers=headers, json=payload)
        response.raise_for_status()
        
        return_data = response.json()
        event_link = return_data.get("htmlLink", "Link indisponível")
        
        return f"Sucesso! Evento criado no Google Calendar. Link: {event_link}"
        
    except requests.exceptions.HTTPError as e:
        error = e.response.json()
        message_error = error.get('error', {}).get('message', str(e))
        
        return f"Falha ao criar o evento no Google Calendar. Erro da API: {message_error}"
    except Exception as e:
        return f"Erro interno ao tentar comunicar com o Google Calendar: {str(e)}"
    
    

@mcp.tool()
def save_google_drive():
    """Funcao responsavel por salvar o arquivo no google drive"""
    return "Funcionalidade de salvar no Google Drive ainda não implementada."


if __name__ == "__main__":
    mcp.run(transport="http", port=8000)