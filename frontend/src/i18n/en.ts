import { TranslationMessages } from 'react-admin';
import englishMessages from 'ra-language-english';

const customEnglishMessages: TranslationMessages = {
    ...englishMessages,
    portal: {
        configuration: "Settings",
        language: "Language",
        theme: {
            type: "Theme type",
            light: "Light",
            dark: "Dark"
        }
    },
    resources: {
        commands: {
            users_name: "Users",
            groups_name: "Groups",
            access_name: "Accesses",
            repository_name: "Repositories"
        },
        users: {
            name: "Users",
            add_title: "Add new user",
            edit_title: "Edit user entry",
            fields: {
                login: "Login",
                name: "Username",
                password: "Password",
                group: "Group",
                role: "Role",
                blocked: "User blocked",
                description: "Description"

            }
        },
        groups: {
            name: "Groups",
            add_title: "Add new group",
            edit_title: "Edit group",
            fields: {
                name: "Username",
                description: "Description"

            }
        },
        accesses:{
            name: "Accesses",
            add_title: "Add new access",
            edit_title: "Edit access",
            fields: {
                name: "Access name",
                owner_id: "User",
                resource_type: "Resource type",
                resource_name: "Repositry name",
                action:"Allowed action",
                disabled:"Disabled"
            }
        }
    }
};

export default customEnglishMessages;
