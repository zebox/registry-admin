
import {
    required,
    AutocompleteInput,
    Create,
    NotFound,
    TextInput,
    SimpleForm,
    ReferenceInput,
    BooleanInput,
    useTranslate,
    usePermissions
} from 'react-admin';
import { RepositoryAutocomplete } from '../components/RepositoryAutocompleteField';
import RepositoryAction from "../components/RepositoryAction";
import { requirePermission } from '../helpers/Helpers';
import { UserAccessSelector } from '../components/UserAccessSelector';

export const AccessCreate = (props: any) => {
    const { source, ...rest } = props;
    const translate = useTranslate();
    const { permissions } = usePermissions();

    return (requirePermission(permissions, 'admin') ?
        <Create title={translate('resources.accesses.add_title')}>
            <SimpleForm>
                <TextInput sx={{ width: "30%" }} label={translate('resources.accesses.fields.name')} source="name"
                                validate={required()}/>
                   {/*  <ReferenceInput source="owner_id" reference="users"
                                    label={translate('resources.accesses.fields.owner_id')}>
                        <AutocompleteInput sx={{width: "30%"}} optionText="name" optionValue="id"
                                           label={translate('resources.accesses.fields.owner_id')}
                                           validate={required()}/>
                    </ReferenceInput> */}
                     <UserAccessSelector source="owner_id" label={translate('resources.accesses.fields.owner_id')}/>
                <RepositoryAutocomplete source="resource_name" {...rest} validate={required()} />
                <RepositoryAction source="action" validate={required()} defaultValue="pull" />
                <TextInput
                    label={translate('resources.accesses.fields.resource_type')}
                    source="type"
                    defaultValue={"repository"}
                    disabled
                />
                <BooleanInput label={translate('resources.accesses.fields.disabled')} source="disabled" />
            </SimpleForm>

        </Create> : <NotFound />
    )
};

export default AccessCreate;