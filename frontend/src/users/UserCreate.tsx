import {
    BooleanInput,
    Create,
    PasswordInput,
    TextInput,
    ReferenceInput,
    SelectInput,
    SimpleForm,
    useTranslate,
    usePermissions,
    NotFound,
    required
} from 'react-admin';
import {requirePermission} from '../helpers/Helpers';
import {RoleList} from "./UsersList";

export const UserCreate = (props:any) => {
    const translate = useTranslate();
    const {source, ...rest} = props;
    const {permissions} = usePermissions();

    return (requirePermission(permissions, 'admin') ?
            <Create title={translate('resources.users.add_title')}>
                <SimpleForm>
                    <TextInput label={translate('resources.users.fields.login')} source="login" {...rest} validate={required()}/>
                    <TextInput label={translate('resources.users.fields.name')} source="name" autoComplete='off' {...rest} validate={required()}/>
                    <PasswordInput label={translate('resources.users.fields.password')} source="password"
                                   autoComplete="new-password" {...rest} validate={required()}/>
                    <ReferenceInput source="group" reference="groups">
                        <SelectInput label={translate('resources.users.fields.group')} 
                        emptyValue={""} 
                        emptyText='' 
                        defaultValue={1}
                        optionText="name" optionValue="id"
                        {...rest} validate={required()}
                         />
                    </ReferenceInput>
                    <SelectInput
                        label={translate('resources.users.fields.role')}
                        source="role"
                        defaultValue={"user"}
                        emptyValue={""}
                        choices={RoleList}
                        {...rest} validate={required()}
                        />
                    <BooleanInput label={translate('resources.users.fields.blocked')} source="blocked"/>
                    <TextInput label={translate('resources.users.fields.description')} source="description"
                               autoComplete='off' fullWidth/>
                </SimpleForm>
            </Create> : <NotFound/>
    )
};

export default UserCreate;