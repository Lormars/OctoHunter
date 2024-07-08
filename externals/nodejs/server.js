// Filename: parserServer.js
const express = require('express');
const bodyParser = require('body-parser');
const esprima = require('esprima');
const estraverse = require('estraverse');  // You need to install this package
const app = express();
const port = 9999;

app.use(bodyParser.text({ type: 'text/plain', limit: '50mb'}));

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

app.listen(port, () => {
    console.log(`Esprima parser server listening at http://localhost:${port}`);
});
