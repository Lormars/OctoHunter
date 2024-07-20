// Filename: parserServer.js
const express = require('express');
const bodyParser = require('body-parser');
const esprima = require('esprima');
const estraverse = require('estraverse');  // You need to install this package
const puppeteer = require('puppeteer');
const app = express();
const port = 9999;

app.use(bodyParser.text({ type: 'text/plain', limit: '50mb'}));
app.use(bodyParser.json());

function analyzeASTForInsecurePostMessage(ast) {
    let findings = [];
    estraverse.traverse(ast, {
        enter: function (node) {
            if (node.type === 'CallExpression' &&
                node.callee.type === 'MemberExpression' &&
                node.callee.property.name === 'addEventListener' &&
                node.arguments.length > 1 &&
                node.arguments[0].value === 'message') {
                const handler = node.arguments[1];
                if (handler.type === 'FunctionExpression' || handler.type === 'ArrowFunctionExpression') {
                    const originChecked = handler.body.body.some(statement =>
                        statement.type === 'IfStatement' &&
                        statement.test.type === 'BinaryExpression' &&
                        statement.test.left.type === 'MemberExpression' &&
                        statement.test.left.property.name === 'origin'
                    );
                    if (!originChecked) {
                        findings.push({ type: 'InsecurePostMessage', location: node.loc.start });
                    }
                }
            }
        }
    });
    return findings;
}

app.post('/parse', (req, res) => {
    try {
        const ast = esprima.parseScript(req.body, { tolerant: true, loc: true });
        const issues = analyzeASTForInsecurePostMessage(ast);
        if (issues.length > 0) {
            res.json({ issues });
        } else {
            res.json({ message: "No insecure postMessage listeners found." });
        }
    } catch (error) {
        res.status(400).json({ error: error.message });
    }
});

app.post('/status', async (req, res) => {
    const url = req.body.url;
    if (!url) {
        return res.status(400).json({ error: "URL is required" });
    }

    let browser;
    try {
        browser = await puppeteer.launch({ headless: true, args: ['--no-sandbox', '--disable-setuid-sandbox'] });
        const page = await browser.newPage();

        await page.setUserAgent('Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36');
        await page.setExtraHTTPHeaders({
            'Accept-Language': 'en-US,en;q=0.9',
            'Accept-Encoding': 'gzip, deflate, br'
        });

        // Set viewport to a common screen size
        await page.setViewport({ width: 1280, height: 800 });

        // Navigate to a page to set cookies
        await page.goto(url, { waitUntil: 'networkidle2' });

        const response = await page.goto(url, { waitUntil: 'networkidle2' });
        const statusCode = response.status();

        await browser.close();
        res.json({ statusCode });
    } catch (error) {
        if (browser) {
            await browser.close();
        }
        res.status(500).json({ error: error.message });
    }
});

app.listen(port, () => {
    console.log(`Esprima parser server listening at http://localhost:${port}`);
});
