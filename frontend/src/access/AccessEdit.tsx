
import {
    AutocompleteInput,
    BooleanInput,
    Edit,
    NotFound,
    TextInput,
    SimpleForm,
    ReferenceInput,
    useTranslate,
    usePermissions,
    required
} from 'react-admin';
import {RepositoryAutocomplete} from "../components/RepositoryAutocompleteField";
import RepositoryAction from "../components/RepositoryAction";
import {requirePermission} from '../helpers/Helpers';
import { UserAccessSelector } from '../components/UserAccessSelector';

const AccessEdit = (props:any) => {
    const { source, ...rest } = props;
    const translate = useTranslate();
    const {permissions} = usePermissions();

    return (requirePermission(permissions, 'admin') ?
            <Edit title={translate('resources.groups.edit_title')}>
                <SimpleForm>
                    <TextInput sx={{width: "30%"}} label={translate('resources.accesses.fields.name')} source="name"
                               validate={required()}/>
                    {/* <ReferenceInput source="owner_id" reference="users">
                        <AutocompleteInput sx={{width: "30%"}} optionText="name" optionValue="id"
                                           label={translate('resources.accesses.fields.owner_id')}
                                           validate={required()}/>
                    </ReferenceInput> */}
                    <UserAccessSelector source="owner_id" label={translate('resources.accesses.fields.owner_id')}/>
                    <RepositoryAutocomplete source="resource_name" validate={required()} {...rest} />
                    <RepositoryAction source="action"/>
                    <TextInput
                        label={translate('resources.accesses.fields.resource_type')}
                        source="type"
                        defaultValue={"repository"}
                        disabled
                    />
                    <BooleanInput label={translate('resources.accesses.fields.disabled')} source="disabled"/>
                </SimpleForm>
            </Edit> : <NotFound/>
    )
};

export default AccessEdit;