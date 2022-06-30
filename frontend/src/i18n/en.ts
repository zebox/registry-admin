import { TranslationMessages } from 'react-admin';
import englishMessages from 'ra-language-english';

const customEnglishMessages: TranslationMessages = {
    ...englishMessages,
    resources: {
        users:{
            name:"Users",
            edit_title:"Edit user entry",
            fields: {
                login:"Login",
                name:"Username",
                password:"Password",
                group:"Group",
                role:"Role",
                blocked:"User blocked",
                description:"Description"
            
            }
        }
    }
};

export default customEnglishMessages;
