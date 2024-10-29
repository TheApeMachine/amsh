/** @jsx jsx */
import { jsx } from '@/lib/template';
import { useEffect, useState } from 'react';
import { Schema, DocCollection, Text } from 'affine';
import { AffineSchemas, AffineEditorContainer } from 'affine';
import { IndexeddbPersistence } from 'affine';

const YouiBlocksuite = () => {
    const [schema, setSchema] = useState<Schema | null>(null);
    const [collection, setCollection] = useState<DocCollection | null>(null);
    const [doc, setDoc] = useState<any>(null);
    const [editor, setEditor] = useState<AffineEditorContainer | null>(null);

    useEffect(() => {
        // Initialize schema, collection, and document
        const newSchema = new Schema().register(AffineSchemas);
        const newCollection = new DocCollection({ schema: newSchema });
        newCollection.meta.initialize();
        const newDoc = newCollection.createDoc();

        // Setting up persistence
        new IndexeddbPersistence('provider-demo', newDoc.spaceDoc);

        // Load and create initial blocks
        newDoc.load(() => {
            const pageBlockId = newDoc.addBlock('affine:page', {
                title: new Text('Test'),
            });
            newDoc.addBlock('affine:surface', {}, pageBlockId);
            const noteId = newDoc.addBlock('affine:note', {}, pageBlockId);
            newDoc.addBlock(
                'affine:paragraph',
                { text: new Text('Hello World!') },
                noteId
            );
        });

        // Initialize editor
        const newEditor = new AffineEditorContainer();
        newEditor.doc = newDoc;

        // Set state
        setSchema(newSchema);
        setCollection(newCollection);
        setDoc(newDoc);
        setEditor(newEditor);
    }, []);

    return (
        <div style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            border: 'var(--youi-border-width, 2px) solid var(--youi-muted, #ccc)',
            borderRadius: 'var(--youi-radius, 0.25rem)',
            width: '100%',
            height: '100%',
        }}>
            <style>
                {`
                :host {
                    --youi-unit: 1rem;
                    --youi-unit-2: 2rem;
                    --youi-radius: 0.25rem;
                    --youi-border-width: 2px;
                    --youi-muted: #ccc;
                }

                affine-editor-container {
                    width: 100%;
                    height: 100%;
                }
                `}
            </style>
            {editor && <affine-editor-container />}
        </div>
    );
};

export default YouiBlocksuite;
