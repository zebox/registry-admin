import * as React from "react";
import { useParams } from 'react-router-dom';
import { useGetOne, useRedirect, useTranslate, useRecordContext, Title,TextField, Loading, Show, SimpleShowLayout } from 'react-admin';
import { Card, Stack, Typography } from '@mui/material';

/**
 * Fetch a repository entry from the API and display it
 */

const RepositoryShowView = ()=>{
    const record = useRecordContext(); // this component is rendered in the /repository/catalog/:repository_name path
    const redirect = useRedirect();
    const translate = useTranslate();
    const { data, isLoading } = useGetOne(
        '/registry/catalog',
        { id: record.repository_name},
        // redirect to the list if the book is not found
        { onError: () => redirect('/registry/catalog') }
    );
    if (isLoading) { return <Loading />; }
    return (
        <div>
            <Title title={translate('resources.repository.title')}/>
            <Card>
                <Stack spacing={1}>
                    <div>
                        <Typography variant="caption" display="block">Title</Typography>
                        <Typography variant="body2">{data.repository_name}</Typography>
                    </div>
                    <div>
                        <Typography variant="caption" display="block">Publication Date</Typography>
                        <Typography variant="body2">{new Date(data.timestamp).toDateString()}</Typography>
                    </div>
                </Stack>
            </Card>
        </div>
    );
}
const RepositoryShow = () => {
    const record = useRecordContext(); // this component is rendered in the /repository/catalog/:repository_name path
    return <Show queryOptions={{ meta: { foo: 'bar' }}}>
     <SimpleShowLayout>
            <TextField source="repository_name" />
     </SimpleShowLayout>
    </Show>
};

export default RepositoryShow;