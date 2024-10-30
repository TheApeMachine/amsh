import { jsx } from '@/lib/template';
import { Schema, DocCollection, Text } from 'affine';
import { AffineSchemas, AffineEditorContainer } from 'affine';
import { IndexeddbPersistence } from 'affine';

const YouiBlocksuite = () => {
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

    return jsx(newEditor, {});
};

export default YouiBlocksuite;
