import uuid
from django.db import models

class Item(models.Model):
    id = models.UUIDField(primary_key=True, editable=False, default=uuid.uuid4, null=False)
    isbn = models.CharField(max_length=64)
    title = models.CharField(max_length=1024)
    author = models.CharField(max_length=1024)
    price = models.DecimalField(decimal_places=2, max_digits=6)
    
    
                          

