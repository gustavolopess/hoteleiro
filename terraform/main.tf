terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }
  }

  backend "s3" {
    bucket = "hoteleiro-bot2"
    key = "terraform.tfstate"
  }

  required_version = ">=1.2.0"
}

provider "aws" {
  region = "us-east-2"
}

resource "aws_instance" "hotelier_server" {
  ami           = "ami-0beaa649c482330f7"
  instance_type = "t2.micro"

  tags = {
    Name   = "HotelierInstance"
    Type   = "t2.micro"
    Cretor = "terraform"
  }
}

output "my-public-ip" {
    value = aws_instance.hotelier_server.public_ip
}
