import * as React from "react";
import { useParams, useLocation } from 'react-router-dom';
import {  useDelete, useTranslate, Datagrid, useRecordContext, Title, ListBase, ListToolbar, Pagination, TextField, Loading } from 'react-admin';
import { Card, CardContent, Typography, Button } from '@mui/material';

import { ConvertUnixTimeToDate, ParseSizeToReadable, SearchFieldTranslated } from "../helpers/Helpers";
import { SizeFieldReadable } from "./RepositoryList";
import ImageConfigPage from './ImageConfig';

/**
 * Fetch a repository entry from the API and display it
 */

export const repositoryBaseResource = 'registry/catalog';

const RepositoryShow = () => {
    const translate = useTranslate();

    return <TagList title={translate('resources.repository.tag_list_title')}>
        <Datagrid bulkActionButtons={false}>
            <TextField source="tag" />
            <TagDescription source="digest" />
            <DateFieldFormatted source="timestamp" />
            <SizeFieldReadable source="size" />
            <TagDeleteButton />
            <ShowImageDetail />
        </Datagrid>
    </TagList>
}


const TagList = ({ children, actions, filters, title, ...props }: any) => {
    const { id } = useParams();
    return (
        <ListBase filter={{ repository_name: id }} queryOptions={{ meta: { group_by: "none" } }}>
            <Title title={id} />
            <ListToolbar
                filters={SearchFieldTranslated()}
            />
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
    return (
        <>
            <ImageConfigPage record={record} />
        </>
    )
}

const TagDeleteButton = ({ source }: any) => {
    const record = useRecordContext();
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


    return <Button onClick={() => deleteTag()}>DELETE</Button>

}
const DateFieldFormatted = ({ source }: any) => {
    const record = useRecordContext();
    return record && <>{ConvertUnixTimeToDate(record[source])}</>
}

export default RepositoryShow;