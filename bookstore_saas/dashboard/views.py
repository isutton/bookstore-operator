import uuid
from django.shortcuts import render
from django.db import models

class Item(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    isbn = models.CharField(max_length=256)
    title = models.CharField(max_length=1024)
    author = models.CharField(max_length=1024)
    price = models.DecimalField()
