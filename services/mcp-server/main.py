from domain.google_calendar import CreateSchedule
import requests
from mcp.server.fastmcp import FastMCP


mcp = FastMCP("bioma-mcp-server")

urls = {
    "google_calendar": "https://www.googleapis.com/calendar/v3/calendars/primary/events",
}


@mcp.tool()
async def great(name):
    """Saudacao para usuario"""
    return f"Hello {name}, how are you? Im fine thanks!!"

@mcp.tool()
async def sum(a, b) -> int:
    """Soma dois numeros"""
    return a + b

@mcp.tool()
async def schedule_google_calendar(input: CreateSchedule, acess_token: str):
    """Funcao responsavel por agendar compromissos no google calendar"""
    headers = {
        "Authorization": f"Bearer {acess_token}",
        "Content-Type": "application/json"
    }
    
    payload = input.model_dump()
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
    """
    Funcao responsavel por salvar o arquivo no google drive 
    """

if __name__ == "__main__":
    mcp.run(transport="http", port=8000)