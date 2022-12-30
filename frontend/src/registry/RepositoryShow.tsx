import * as React from "react";
import { useParams } from 'react-router-dom';
import {
    useDelete,
    usePermissions,
    useTranslate,
    useRefresh,
    useNotify,
    Datagrid,
    useRecordContext,
    Title,
    ListBase,
    ListToolbar,
    Pagination,
    TextField,
    Loading,
    Confirm
} from 'react-admin';
import { Card, CardContent, Typography, Button } from '@mui/material';
import { ConvertUnixTimeToDate, SearchFieldTranslated } from "../helpers/Helpers";
import { SizeFieldReadable } from "./RepositoryList";
import ImageConfigPage from './ImageConfig';
import InfoIcon from '@mui/icons-material/Info';
import { requirePermission } from '../helpers/Helpers';
import CopyToClipboard from "../components/ClipboardCopy/CopyToClipboard";
/**
 * Fetch a repository entry from the API and display it
 */

export const repositoryBaseResource = 'registry/catalog';
let isErrorShow: Boolean = false;

const RepositoryShow = () => {
    const translate = useTranslate();
    const { permissions } = usePermissions();

    return <TagList title={translate('resources.repository.tag_list_title')}>
        <Datagrid bulkActionButtons={false}>
            <TagFiled source="tag" label={translate('resources.repository.fields.tag')} />
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

const TagFiled = ({ source }: any): React.ReactElement => {
    const record = useRecordContext();
    return (
        <>
            <TextField source="tag" />
            <CopyToClipboard content={`${record['repository_name']}:${record[source]}`} />
        </>
    )
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

const TagDeleteButton = (props: any) => {
    const record = useRecordContext();
    const translate = useTranslate();
    const refresh = useRefresh();
    const notify = useNotify();

    const [deleteOne, { isLoading, error }] = useDelete();
    const [open, setOpen] = React.useState(false);
    const [deleteDisable, setDeleteDisable] = React.useState(false);

    const handleDialogClose = () => {
        setOpen(false);
    }

    const deleteTag = () => {
        setDeleteDisable(true);
        isErrorShow = false;
        deleteOne(
            repositoryBaseResource,
            { id: record["tag"], previousData: record, meta: { name: record["repository_name"], digest: record["digest"] } }
        ).finally(() => {
            // wait until changing sync
            setTimeout(() => {
                setDeleteDisable(false);
                refresh();
            }, 2000);

        })
        setOpen(false);
    }

    if (isLoading) return <Loading />

    if (!isErrorShow && error) {
        isErrorShow = true;
        const err = error as Error;
        notify(
            typeof error === 'string'
                ? error
                :
                typeof err === 'undefined' || err.message
                    ? err.message
                    : 'ra.notification.http_error',
            {
                type: 'warning',
                messageArgs: {
                    _:
                        typeof error === 'string'
                            ? error
                            : err && err.message
                                ? err.message
                                : undefined,
                },
            }
        )

    }


    return <React.Fragment>
        <Button onClick={() => setOpen(true)} disabled={deleteDisable}>{translate('ra.action.delete')}</Button>
        <Confirm
            isOpen={open}
            loading={isLoading}
            title='ra.message.delete_title'
            content='ra.message.delete_content'
            translateOptions={{
                name: record["repository_name"],
                id: record["tag"]
            }}
            onConfirm={deleteTag}
            onClose={handleDialogClose}
        />

    </React.Fragment>

}
const DateFieldFormatted = ({ source }: any) => {
    const record = useRecordContext();
    return record && <>{ConvertUnixTimeToDate(record[source])}</>
}

export default RepositoryShow;