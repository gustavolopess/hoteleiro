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
    region = "us-east-2"
  }

  required_version = ">=1.2.0"
}

provider "aws" {
  region = "us-east-2"
}

resource "aws_instance" "hotelier_server" {
  ami           = "ami-0b299847e139e681b"
  instance_type = "t2.micro"
  security_groups = [ "HotelierAllowSSH" ]

  tags = {
    Name   = "HotelierInstance"
    Type   = "t2.micro"
    Cretor = "terraform"
  }
}

output "public_dns" {
    value = aws_instance.hotelier_server.public_dns
}
