import * as React from "react";
import { useParams, useLocation } from 'react-router-dom';
import { useGetOne, useGetList, useRedirect, useTranslate, Datagrid, useRecordContext, useListController, Title, ListBase, ListToolbar, Pagination, TextField, ShowContextProvider, RecordContextProvider, Loading, ListContextProvider } from 'react-admin';
import { Box, Card, CardContent, Stack, Typography } from '@mui/material';
import { ConvertUnixTimeToDate, ParseSizeToReadable } from "../helpers/Helpers";
import { SizeFieldReadable } from "./RepositoryList";

/**
 * Fetch a repository entry from the API and display it
 */

const repositoryBaseResource = 'registry/catalog';

const RepositoryShow = () => {
    const translate = useTranslate();

    return <TagList title={translate('resources.repository.tag_list_title')}>
        <Datagrid bulkActionButtons={false}>
            <TextField source="tag" />
            <TagDescription source="digest"/>
            <DateFieldFormatted source="timestamp" />
            <SizeFieldReadable source="size" />
        </Datagrid>
    </TagList>
}


const TagList = ({ children, actions, filters, title, ...props }: any) => {
    const { id } = useParams();
    return (
        <ListBase filter={{ repository_name: id }} queryOptions={{ meta: { group_by: "none" } }}>
            <Title title={id} />
            <ListToolbar
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
const DateFieldFormatted = ({ source }: any) => {
    const record = useRecordContext();
    return record && <>{ConvertUnixTimeToDate(record[source])}</>
}

export default RepositoryShow;