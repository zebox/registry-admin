import * as React from "react";
import { useParams } from 'react-router-dom';
import { useDelete, usePermissions, useTranslate, Datagrid, useRecordContext, Title, ListBase, ListToolbar, Pagination, TextField, Loading } from 'react-admin';
import { Card, CardContent, Typography, Button } from '@mui/material';

import { ConvertUnixTimeToDate, SearchFieldTranslated } from "../helpers/Helpers";
import { SizeFieldReadable } from "./RepositoryList";
import ImageConfigPage from './ImageConfig';
import InfoIcon from '@mui/icons-material/Info';
import {requirePermission} from '../helpers/Helpers';
/**
 * Fetch a repository entry from the API and display it
 */

export const repositoryBaseResource = 'registry/catalog';

const RepositoryShow = () => {
    const translate = useTranslate();
    const { permissions } = usePermissions();

    return <TagList title={translate('resources.repository.tag_list_title')}>
        <Datagrid bulkActionButtons={false}>
            <TextField source="tag" label={translate('resources.repository.fields.tag')} />
            <TagDescription source="digest" label={translate('resources.repository.fields.digest')} />
            <DateFieldFormatted source="timestamp" label={translate('resources.repository.fields.date')} />
            <SizeFieldReadable source="size" label={translate('resources.repository.fields.size')} />
            {requirePermission(permissions, 'admin') ?
                <>
                    <TagDeleteButton />
                    <ShowImageDetail />
                </> : null}
        </Datagrid>
    </TagList>
}


const TagList = ({ children, actions, filters, title, ...props }: any) => {
    const { id } = useParams();
    const translate = useTranslate();
    return (
        <ListBase filter={{ repository_name: id }} queryOptions={{ meta: { group_by: "none" } }}>
            <Title title={id} />
            <ListToolbar filters={SearchFieldTranslated(translate)} />
            <Card >
                {children}
            </Card>
            <Pagination />
        </ListBase >
    );
}

const TagDescription = ({ source }: any) => {
    const record = useRecordContext();
    const translate = useTranslate();
    return <Card sx={{ minWidth: 275 }}>
        <CardContent>
            <Typography sx={{ fontSize: 14 }} color="text.secondary" gutterBottom>
                {translate('resources.repository.pull_counter')} {record["pull_counter"]}
            </Typography>
            <Typography sx={{ fontSize: 14 }} color="text.secondary" gutterBottom>
                {translate('resources.repository.tag_digest')} <i>{record[source]}</i>
            </Typography>
            <Typography sx={{ fontSize: 14 }} color="text.secondary" gutterBottom>
                {translate('resources.repository.tag_media_type')} <i>{record["raw"].mediaType}</i>
            </Typography>
        </CardContent>
    </Card>
}

const ShowImageDetail = ({ source }: any) => {
    const record = useRecordContext();
    const translate = useTranslate();
    const [open, setOpen] = React.useState(false);

    const handleClickOpen = () => {
        setOpen(true);
    }
    return (
        <>{open ?
            <ImageConfigPage record={record} isOpen={open} handleShowFn={setOpen} />
            :
            <Button variant="outlined" onClick={handleClickOpen}>
                {translate('resources.repository.fields.details')}
                <InfoIcon />
            </Button>
        }
        </>
    )
}

const TagDeleteButton = ({ source }: any) => {
    const record = useRecordContext();
    const translate = useTranslate();

    const [deleteOne, { isLoading, error }] = useDelete();

    const deleteTag = () => {
        deleteOne(
            repositoryBaseResource,
            { id: record["tag"], previousData: record, meta: { name: record["repository_name"], digest: record["digest"] } }
        );

    }

    if (isLoading) return <Loading />
    if (error) {
        console.error(error);
    }


    return <Button onClick={() => deleteTag()}>{translate('ra.action.delete')}</Button>

}
const DateFieldFormatted = ({ source }: any) => {
    const record = useRecordContext();
    return record && <>{ConvertUnixTimeToDate(record[source])}</>
}

export default RepositoryShow;