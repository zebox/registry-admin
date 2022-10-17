import SyncRepo from '@mui/icons-material/Sync';
import { Box, Typography,Card,CardContent } from '@mui/material';
import { useState } from 'react';

import { SearchFieldTranslated, ParseSizeToReadable } from '../helpers/Helpers';

import {
    Button,
    ExportButton,
    ShowButton,
    DeleteButton,
    List,
    Loading,
    Datagrid,
    TopToolbar,
    TextField,
    useNotify,
    useDataProvider,
    useTranslate,
    useRecordContext,
    useCreatePath,
    useRedirect
} from 'react-admin';


const EmptyList = () => {
    const translate = useTranslate();
    return (
        <Box textAlign="center" m={1}>
            <Card>
            <CardContent>
                    <Typography variant="h4" paragraph>
                        {translate('resources.repository.message_empty_page')}
                    </Typography>
                    <Typography variant="body1">
                        {translate('resources.repository.message_sync_repo')}
                    </Typography>
                    <SyncButton/>
            </CardContent>
            </Card>
    </Box>
    )
};

const SyncButton = () =>{
    const dataProvider = useDataProvider();
    const notify = useNotify();
    const [isLoading, setLoading] = useState(false)
    const translate = useTranslate();

    const doRepoSync = () => {
        setLoading(true);
        dataProvider.getList('registry/sync', { 
            pagination: { page: 1, perPage: 10 },
            sort: { field: 'id', order: 'DESC' },
            filter:{}
        })
        .then(({ data }) => {
            setLoading(false);
            notify(translate('resources.repository.message_syncing_repo', { type: 'success' }))
        })
        .catch(error => {
            setLoading(false);

            if (error.body.message.includes("repository sync currently running")) {
                notify(translate('resources.repository.message_repo_syncing_running'), { type: 'error' })
                return
            }

            notify(translate('resources.repository.message_error_syncing_repo')+` ${error.message}`, { type: 'error' })
        })
        setLoading(false);
    }

    if (isLoading) { 
        return <Loading />; 
    }

    return (
        <Button
        onClick={() => {doRepoSync()}}
        label={translate('resources.repository.button_sync')}
        >
            <SyncRepo/>
        </Button>
    )

}

const RepositoryShowButton = ()=> {
    const record = useRecordContext();
    const createPath = useCreatePath();
    const redirect = useRedirect();

    const doShow = () =>{
       console.log("click...");
       redirect(createPath({ resource: 'registry/catalog', type: 'show', id: record.repository_name }))
    }
    if (record) {
        record.id=record.repository_name;
    }
   //  return <p onClick={()=>doShow()}>fsdf</p>
   return record && <ShowButton record={record} />
}

const RepositoryActions = () => (
    <TopToolbar>
        <ExportButton/>
        <SyncButton/>
    </TopToolbar>
);

const RepositoryList = (props:any) => (
    <List 
        {...props}
        empty={<EmptyList/>}
        actions={<RepositoryActions/>}
        title={useTranslate()(`resources.commands.repository_name`)}
        sort={{ field: 'repository_name', order: 'ASC' }}
        perPage={25}
        filters={SearchFieldTranslated()}
    >
        <Datagrid bulkActionButtons={false}>
           {/*  <TextField source="id" /> */}
            <TextField source="repository_name" />
            <SizeFieldReadable source="size" />
            <RepositoryShowButton/>
        </Datagrid>
    </List>
);

export const SizeFieldReadable= ({ source }: any) => {
    const record = useRecordContext();

    return record ? (
        <>
            {ParseSizeToReadable(record[source],2)}
        </>
    ) : null;
}

export default RepositoryList;