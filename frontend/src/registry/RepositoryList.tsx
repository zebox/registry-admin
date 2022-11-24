import React from 'react';
import SyncRepo from '@mui/icons-material/Sync';
import { Box, Typography, Card, Container, CssBaseline, CardContent, Tooltip } from '@mui/material';
import { useEffect, useState } from 'react';

import { SearchFieldTranslated, ParseSizeToReadable } from '../helpers/Helpers';

import {
    Button,
    ExportButton,
    ShowButton,
    List,
    Loading,
    Datagrid,
    TopToolbar,
    TextField,
    useNotify,
    useDataProvider,
    useTranslate,
    useRecordContext,
    usePermissions
} from 'react-admin';
import { requirePermission } from '../helpers/Helpers';


const EmptyList = () => {
    const translate = useTranslate();

    return (
        <React.Fragment >
            <CssBaseline />
            <Container maxWidth="sm">
                <Box textAlign="center" m={1}>
                    <Card>
                        <CardContent>
                            <Typography variant="h4" paragraph>
                                {translate('resources.repository.message_empty_page')}
                            </Typography>
                            <Typography variant="body1">
                                {translate('resources.repository.message_sync_repo')}
                            </Typography>
                            <SyncButton />
                        </CardContent>
                    </Card>
                </Box>
            </Container>
        </React.Fragment>

    )
};

const SyncButton = () => {
    const dataProvider = useDataProvider();
    const notify = useNotify();
    const [isLoading, setLoading] = useState(false)
    const translate = useTranslate();
    const [isAdmin, setIsAdmin] = useState(false);
    const { permissions } = usePermissions();

    useEffect(() => {
        setIsAdmin(requirePermission(permissions, 'admin'));
    }, [permissions]);

    const doRepoSync = () => {
        setLoading(true);
        dataProvider.getList('registry/sync', {
            pagination: { page: 1, perPage: 10 },
            sort: { field: 'id', order: 'DESC' },
            filter: {}
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

                notify(translate('resources.repository.message_error_syncing_repo') + ` ${error.message}`, { type: 'error' })
            })
        setLoading(false);
    }

    if (isLoading) {
        return <Loading />;
    }

    return (isAdmin ?

        <Button
            onClick={() => { doRepoSync() }}
            label={translate('resources.repository.button_sync')}
        >
            <Tooltip title={translate('resources.repository.message_sync_about')}>
                <SyncRepo />
            </Tooltip>
        </Button>
        : null
    )

}

const RepositoryShowButton = () => {
    const record = useRecordContext();
    if (record) {
        record.id = record.repository_name;
    }
    return record && <ShowButton record={record} />
}

const RepositoryActions = () => {

    return (
        <TopToolbar>
            <ExportButton />
            <SyncButton />
        </TopToolbar>
    )
};

const RepositoryList = (props: any) => {
    const translate = useTranslate();
    return (
        <List
            {...props}
            empty={<EmptyList />}
            actions={<RepositoryActions />}
            title={translate(`resources.commands.repository_name`)}
            sort={{ field: 'repository_name', order: 'ASC' }}
            perPage={25}
            filters={SearchFieldTranslated(translate)}
        >
            <Datagrid bulkActionButtons={false}>
                <TextField source="repository_name" label={translate('resources.repository.fields.name')} />
                <SizeFieldReadable source="size" label={translate('resources.repository.fields.size')} />
                <RepositoryShowButton />
            </Datagrid>
        </List>
    )
};

export const SizeFieldReadable = ({ source }: any) => {
    const record = useRecordContext();

    return record ? (
        <>
            {ParseSizeToReadable(record[source], 2)}
        </>
    ) : null;
}

export default RepositoryList;