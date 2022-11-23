
import {
    AutocompleteInput,
    List,
    Datagrid,
    TextField,
    EditButton,
    SelectInput,
    ReferenceInput,
    ReferenceField,
    useRecordContext,
    useTranslate,
    usePermissions,
    NotFound
} from 'react-admin';
import Button from '@mui/material/Button';
import { Link } from 'react-router-dom';
import AccessIcon from '@mui/icons-material/LockOpen';
import Tooltip from '@mui/material/Tooltip'
import { DisabledField } from '../components/DisabledField';
import {SearchFieldTranslated} from '../helpers/Helpers'
import {requirePermission} from '../helpers/Helpers';
import { DeleteCustomButtonWithConfirmation } from '../components/DeleteCustomButtonWithConfirmation';

interface IRoleList {
    id: string;
    name: string;
}

export const KEY_SELECTED_USER="ctx_key_selected_user_";

export const RoleList: Array<IRoleList> = [
    { id: 'user', name: 'User' },
    { id: 'manager', name: 'Manager' },
    { id: 'admin', name: 'Admin' }
]

const RepositoryAccessList =()=>{
    const record = useRecordContext();
    const translate=useTranslate();

    return record ? (
        <Tooltip title={translate('resources.accesses.messages.access_tooltip')}>
        <Button
            color="primary"
            component={Link}
            to={{
                pathname: '/access',
                search: `filter=${JSON.stringify({ owner_id: record.id })}`,
            }}
        >
            <AccessIcon/>
        </Button>
        </Tooltip>
    ) : null;

}


const UserList = (props:any) => {
    const translate=useTranslate();
    const {permissions} = usePermissions();

    return(requirePermission(permissions,'admin') ?
    <List
        sort={{ field: 'name', order: 'ASC' }}
        perPage={25}
        filters={SearchFieldTranslated(translate,[<SelectInput
            source="role"
            defaultValue={"user"}
            emptyValue={null}
            choices={RoleList} />,
            <ReferenceInput source="user_group" reference="groups" label={translate('resources.groups.name')}>
               <AutocompleteInput  optionText="name" optionValue="id" label={translate('resources.groups.name')} />
           </ReferenceInput>])}
    >
        <Datagrid bulkActionButtons={false}>
            <TextField source="id" />
            <TextField source="login" />
            <TextField source="name"  label={translate('resources.groups.fields.name')}/>
            <ReferenceField source="group" reference="groups">
                <TextField source="name" />
            </ReferenceField>
            <TextField source="role" />
            <DisabledField source="blocked" />
            <RepositoryAccessList/>
            <EditButton />
            <DeleteCustomButtonWithConfirmation  source="name" {...props}/>
        </Datagrid>
    </List>:<NotFound/>
)};


export default UserList;