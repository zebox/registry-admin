import { useEffect, useState } from "react";
import {
    List,
    Datagrid,
    ReferenceInput,
    AutocompleteInput,
    TextField,
    EditButton,
    ReferenceField,
    useTranslate,
    usePermissions,
    useRecordContext,
    FilterLiveSearch,
    FilterList,
    FilterListItem,
} from 'react-admin';
import {DeleteCustomButtonWithConfirmation} from '../components/DeleteCustomButtonWithConfirmation';
import {DisabledField} from "../components/DisabledField";
import {requirePermission} from '../helpers/Helpers'
import {Box, Card, CardContent} from '@mui/material';
import WorkspacePremiumIcon from '@mui/icons-material/WorkspacePremium';
import SwapVerticalCircleOutlinedIcon from '@mui/icons-material/SwapVerticalCircleOutlined';

const AccessFilterSidebar = () => {
    const translate = useTranslate();
    return <Box
        sx={{
            display: {
                xs: 'none',
                sm: 'block'
            },
            mt: 8,
            order: -1, // display on the left rather than on the right of the list
            width: '15em',
            marginRight: '0.5em',
        }}
    >
        <Card>
            <CardContent>
                <FilterLiveSearch/>
                <FilterList label={translate('resources.accesses.labels.label_special_permission')}
                            icon={<WorkspacePremiumIcon/>}>
                    <FilterListItem label={translate('resources.accesses.labels.label_for_all_users')}
                                    value={{owner_id: -1000}}/>
                    <FilterListItem label={translate('resources.accesses.labels.label_for_registered_users')}
                                    value={{owner_id: -999}}/>
                </FilterList>
                <FilterList label={translate('resources.accesses.fields.action')}
                            icon={<SwapVerticalCircleOutlinedIcon/>}>
                    <FilterListItem label="- PULL" value={{action: 'pull'}}/>
                    <FilterListItem label="- PUSH" value={{action: 'push'}}/>
                </FilterList>
            </CardContent>
        </Card>
    </Box>
};

const AccessList = (props: any) => {
    const translate = useTranslate();
    const { permissions } = usePermissions();


    return <List
        aside={<AccessFilterSidebar />}
        hasCreate={requirePermission(permissions, 'admin')}
        sort={{ field: 'name', order: 'ASC' }}
        perPage={25}
        filters={[
            <ReferenceInput source="owner_id" reference="users" label={translate('resources.accesses.fields.owner_id')}>
                <AutocompleteInput optionText="name" optionValue="id" label={translate('resources.accesses.fields.owner_id')} />
            </ReferenceInput>
        ]} >
        <Datagrid bulkActionButtons={false} >
            <TextField source="name" label={translate('resources.accesses.fields.name')} />
            <CustomUsersReference label={translate('resources.accesses.fields.owner_id')} />
            <TextField source="type" label={translate('resources.accesses.fields.resource_type')} />
            <TextField source="resource_name" label={translate('resources.accesses.fields.resource_name')} />
            <TextField source="action" label={translate('resources.accesses.fields.action')} />
            <DisabledField source="disabled" label={translate('resources.accesses.fields.disabled')} />
            {requirePermission(permissions, 'admin') ? <>
                <EditButton alignIcon="right" />
                <DeleteCustomButtonWithConfirmation source="name" {...props} />
            </> : null}
        </Datagrid>
    </List>
};



const CustomUsersReference = (props: any) => {
    const translate = useTranslate();
    const record = useRecordContext();
    const [isUserReference, setIsUserREference] = useState(false);

    const specsPermissionRefs = [
        { name: translate('resources.accesses.labels.label_for_all_users') },
        { name: translate('resources.accesses.labels.label_for_registered_users') }
    ];

    useEffect(() => {
        if (record && record.owner_id < 0) {
            setIsUserREference(true)
        }
    }, [record])

    return (
        !isUserReference ?
            <ReferenceField source="owner_id" reference="users" label={translate('resources.accesses.fields.owner_id')}>
                <TextField source="name" />
            </ReferenceField> :
            <>{record.owner_id === -1000 ? specsPermissionRefs[0].name : specsPermissionRefs[1].name}</>
    )
}

export default AccessList;

