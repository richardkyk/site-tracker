# Price Tracker

A serverless application that retrieves the price of a product from a website.

## Usage

To use this application, you need to provide the following parameters:

- `url`: The URL of the website to scrape.
- `selector`: The CSS selector to use to narrow down the elements to scrape.
- `regex`: The regular expression to use to extract the price.

For example, if you want to scrape the price of a product from a website, you can use the following command:

```bash
curl "https://[xxxxx].execute-api.ap-southeast-2.amazonaws.com/?url=https%3A%2F%2Fexample.com&selector=p&regex=%28.*%3F%29%5C."
```

This will return the price of the product as a plain text response.

## Building

To build this application, you can use the following command:

```bash
docker buildx build --platform linux/arm64 --provenance=false -t lambda:price-tracker .
```
This will build a minimal Docker image for the application using the AWS Lambda runtime.

## Deployment
### (Local)
To deploy this application locally, you can use the following command:

```bash
docker run -p 9000:8080 --entrypoint /usr/local/bin/aws-lambda-rie lambda:price-tracker /var/task/bootstrap
```

This will start a local server that listens on port 9000 and serves the application.

### (AWS)
To deploy this application to AWS, you can use the following command:

```bash
terraform -chdir=infra init
```

This will initialize the Terraform configuration.

```bash
terraform -chdir=infra apply
```
This will build and deploy the Docker image to ECR, create an IAM role for Lambda execution, and create a Lambda function.
