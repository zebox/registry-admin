import * as React from "react";
import { useParams, useLocation } from 'react-router-dom';
import { useGetOne, useGetList, useRedirect, useTranslate, useRecordContext, useGetResourceLabel, Title, TextField, ShowContextProvider, RecordContextProvider, Loading, SimpleShowLayout } from 'react-admin';
import { Card, Stack, Typography } from '@mui/material';

/**
 * Fetch a repository entry from the API and display it
 */

const repositoryBaseResource = 'registry/catalog';

const RepositoryShowView = ({ children }: any) => {
    const record = useRecordContext(); // this component is rendered in the /repository/catalog/:repository_name path
    const recordId = useGetResourceLabel();
    const redirect = useRedirect();
    const translate = useTranslate();
    const { data, isLoading } = useGetOne(
        repositoryBaseResource,
        { id: record.repository_name },
        // redirect to the list if the item is not found
        { onError: () => redirect('/registry/catalog') }
    );
    if (isLoading) { return <Loading />; }
    return (
        <div>
            <Title title={translate('resources.repository.title')} />
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

const MyShowView = () => {

    const record = useRecordContext();
    return (record &&
        <div>
            <Title title={record.repository_name} />
            <Card>
                <Stack spacing={1}>
                    <div>
                        <Typography variant="caption" display="block">Title</Typography>
                        <Typography variant="body2">{record.repository_name}</Typography>
                    </div>
                    <div>
                        <Typography variant="caption" display="block">Publication Date</Typography>
                        <Typography variant="body2">{new Date(record.timestamp).toDateString()}</Typography>
                    </div>
                </Stack>
            </Card>
        </div>
    );
}

const GetRepositoryTag = () => {
    const record = useRecordContext();
    // return { id: 123, title: 'Hello world' };
    return <p>fsdfdsfH</p>
}



const RepositoryShow = (props: any) => {
    const { id } = useParams();
    const redirect = useRedirect();

    const { data, isLoading, error } = useGetList(
        repositoryBaseResource,
        {
            filter: { repository_name: id }
        },
        { onError: () => redirect("/"+repositoryBaseResource) }
    );

    if (isLoading) { return <Loading />; }
    if (error) { return <p>ERROR</p>; }

    return (
        <RecordContextProvider value={data} >
            <MyShowView />
        </RecordContextProvider>
    );
}


export const RepositoryTags = () => {
    const record = useRecordContext();
    const location = useLocation();
    // return { id: 123, title: 'Hello world' };
    return <p>fsdfdsfH</p>
}

/* const RepositoryShow = (props:any):React.ReactElement<any,any> => {
    const record = useRecordContext();
    const redirect = useRedirect();

    const onError = (error:any) => {
       console.error(error)
    };

    return (
        <SimpleShowLayout>
            <GetRepositoryTag />
        </SimpleShowLayout>
  
    )
} */



export default RepositoryShow;