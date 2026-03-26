from fastmcp import FastMCP

mcp = FastMCP("bioma-mcp-server")

@mcp.tool
def schedule_google_calendar():
    """
    Funcao responsavel por agendar compromissos no google calendar
    """

@mcp.tool
def save_google_drive():
    """
    Funcao responsavel por salvar o arquivo no google drive 
    """

if __name__ == "__main__":
    mcp.run()