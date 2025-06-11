import express, { Request, Response } from 'express';
import React from 'react';
import ReactDOMServer from 'react-dom/server';
import path from 'path';

const app = express();
const port = 3001;

// Serve the client bundle for hydration
app.use('/client.js', express.static(path.join(__dirname, 'dist', 'client.js')));

app.use(express.json());

interface RenderRequest {
    componentPath: string;
    props: { [key: string]: any };
}

app.post('/render', (req: Request, res: Response) => {
    const { componentPath, props } = req.body as RenderRequest;

    if (!componentPath) {
        return res.status(400).json({ error: 'componentPath is required' });
    }

    try {
        const componentFullPath = path.join(__dirname, '..', 'user-app', componentPath);
        
        // Ensure we don't have stale cache during development
        if (process.env.NODE_ENV !== 'production') {
            delete require.cache[require.resolve(componentFullPath)];
        }
        
        const ComponentModule = require(componentFullPath);
        const Component = ComponentModule.default || ComponentModule;

        if (typeof Component !== 'function') {
            throw new Error(`Failed to load component from ${componentPath}. Make sure it has a default export.`);
        }

        const element = React.createElement(Component, props);
        const html = ReactDOMServer.renderToString(element);

        res.json({ html });
    } catch (error: any) {
        console.error('Error rendering component:', error);
        res.status(500).json({ error: 'Failed to render component', details: error.message });
    }
});

app.listen(port, () => {
    console.log(`Node.js renderer listening on http://localhost:${port}`);
}); 