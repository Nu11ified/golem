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
    layoutPath?: string;
    props: { [key: string]: any };
}

app.post('/render', async (req: Request, res: Response) => {
    const { componentPath, layoutPath, props } = req.body as RenderRequest;

    if (!componentPath) {
        return res.status(400).json({ error: 'componentPath is required' });
    }

    try {
        const componentFullPath = path.join(__dirname, '..', 'user-app', componentPath);
        let layoutFullPath: string | undefined = undefined;
        if (layoutPath) {
            layoutFullPath = path.join(__dirname, '..', 'user-app', layoutPath);
        }

        // Ensure we don't have stale cache during development
        if (process.env.NODE_ENV !== 'production') {
            delete require.cache[require.resolve(componentFullPath)];
            if (layoutFullPath) {
                delete require.cache[require.resolve(layoutFullPath)];
            }
        }

        const PageModule = require(componentFullPath);
        const Page = PageModule.default || PageModule;
        const pageMetadata = PageModule.metadata || {};

        if (typeof Page !== 'function') {
            throw new Error(`Failed to load component from ${componentPath}. Make sure it has a default export.`);
        }

        let element: React.ReactElement;
        let layoutMetadata = {};
        if (layoutFullPath) {
            const LayoutModule = require(layoutFullPath);
            const Layout = LayoutModule.default || LayoutModule;
            layoutMetadata = LayoutModule.metadata || {};
            if (typeof Layout !== 'function') {
                throw new Error(`Failed to load layout from ${layoutPath}. Make sure it has a default export.`);
            }
            element = React.createElement(Layout, props, React.createElement(Page, props));
        } else {
            element = React.createElement(Page, props);
        }

        // Merge metadata: layout takes precedence over page
        const metadata = { ...pageMetadata, ...layoutMetadata };

        const html = ReactDOMServer.renderToString(element);
        res.json({ html, metadata });
    } catch (error: any) {
        console.error('Error rendering component:', error);
        res.status(500).json({ error: 'Failed to render component', details: error.message });
    }
});

app.listen(port, () => {
    console.log(`Node.js renderer listening on http://localhost:${port}`);
}); 