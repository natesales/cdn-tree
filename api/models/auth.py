from pydantic import BaseModel, EmailStr


class User(BaseModel):
    """
    User stores information about a user
    """
    email: EmailStr  # Email address
    password: str  # Hashed password
