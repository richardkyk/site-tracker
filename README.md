# Site Tracker

A serverless application that notifies you when a specific value has changed on a website.

## Usage

To use this application, you need to provide the following parameters:

- `url`: The URL of the website to scrape.
- `selector`: The CSS selector to use to narrow down the elements to scrape.
- `regex`: The regular expression to use to extract the value. This should include capture groups to extract the value.
- `expected`: The expected value. This would be the current price of a product, for example.
- `email`: The email address to send the notification to.

### Price Tracking

For example, if you want to scrape the price of a product from a website, you can use the following command:

```javascript
const url = "https://[xxxxx].execute-api.[region].amazonaws.com/"
// create the resource and store it in dynamoDB
const res = await fetch(url, {
    method: "POST",
    headers: {
        "Content-Type": "application/json"
    },
    body: JSON.stringify({
        url: "https://books.toscrape.com/catalogue/tipping-the-velvet_999/index.html",
        selector: "div.product_main p.price_color",
        regex: "(\\d+\\.\\d+)",
        expected: "53.74", 
        email: "example@example.com"
    })
})
const id = await res.text()

// returns the extracted value
const res2 = await fetch(`${url}?id=${id})
const value = await res2.text()

```

This will return the price of the product as a plain text response.

### Keyword Tracking

For example, if you want to be notified when the keywords from a website suddenly appear:

```javascript
const url = "https://[xxxxx].execute-api.[region].amazonaws.com/"
// create the resource and store it in dynamoDB
const res = await fetch(url, {
    method: "POST",
    headers: {
        "Content-Type": "application/json"
    },
    body: JSON.stringify({
        url: "https://books.toscrape.com/catalogue/tipping-the-velvet_999/index.html",
        selector: "body",
        regex: "(?i)(sale|discount)",
        expected: "", 
        email: "example@example.com"
    })
})
const id = await res.text()

const res2 = await fetch(`${url}?id=${id})
const value = await res2.text()
```

This will return the extracted value as a plain text response.

## Building

To build this application, you can use the following command:

```bash
make all
```
This will build all the Go binaries and create a zip file for deployment.

## Deployment
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
