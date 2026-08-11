from norns import tool


@tool
def greet(name: str) -> str:
    """Greet someone by name."""
    return f"Hello, {name}!"
