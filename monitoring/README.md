# Please read me!

# Product Requirement Document


# Deployment
### clean compile
####  If you use eclipse
    pom.xml > Run As -> Run Configurations...
                     -> goals:  clean compile    , profiles: env
                     -> parameters:   env = dev
### command examples:
    mvn clean package -DskipTests -Ppure-classes